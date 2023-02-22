package redis

import (
	"context"
	"errors"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/gocron"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/go-redsync/redsync/v4"
	"strconv"
	"strings"
)

const frequency = 10

func getKeys(ctx context.Context, keyPatten string) ([]string, error) {
	//keys, cursor, err := GetRedisHelper().Scan(ctx, 0, keyPatten, 10).Result()
	keys, err := GetRedisHelper().Keys(ctx, keyPatten).Result()
	if err != nil {
		return nil, err
	}
	return keys, err
}

func deleteKeys(ctx context.Context, key string, mutex *redsync.Mutex) error {
	// 先加锁
	errLock := LockByMutex(ctx, mutex)
	if errLock != nil {
		return errors.New("lock failed: " + errLock.Error())
	}
	// Redis处理
	errRedis := GetRedisHelper().Del(ctx, key).Err()
	// 在处理错误返回之前解锁
	errUnlock := UnlockByMutex(ctx, mutex)
	if errUnlock != nil {
		return errors.New("unlock failed: " + errUnlock.Error())
	}
	// 返回Redis错误
	if errRedis != nil {
		return errRedis
	}
	return nil
}

func FavoriteMoveToDB() error {
	logger := zap.InitLogger()

	ctx := context.Background()
	keys, err := getKeys(ctx, "video::*::user::*::w")
	if err != nil {
		logger.Errorln(err)
		return err
	}
	for _, key := range keys {
		actionType, err := GetRedisHelper().Get(ctx, key).Result()
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		// 拆分得key
		kSplit := strings.Split(key, "::")
		vid, uid := kSplit[1], kSplit[3]
		videoID, err := strconv.ParseInt(vid, 10, 64)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		userID, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		// 查询是否存在点赞记录
		favorite, err := db.GetFavoriteVideoRelationByUserVideoID(ctx, userID, videoID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		} else if favorite == nil && actionType == "1" {
			// 数据库中没有该点赞记录，且最终状态为点赞，则插入数据库
			video, err := db.GetVideoById(ctx, videoID)
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
			err = db.CreateVideoFavorite(ctx, userID, videoID, int64(video.AuthorID))
			// 插入后，删除Redis中对应记录
			delErr := deleteKeys(ctx, key, favoriteMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
		} else if favorite != nil && actionType == "2" {
			// 数据库中有该点赞记录，且最终状态为取消点赞，则从数据库中删除该记录
			video, err := db.GetVideoById(ctx, videoID)
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
			err = db.DelFavoriteByUserVideoID(ctx, userID, videoID, int64(video.AuthorID))
			// 插入后，删除Redis中对应记录
			delErr := deleteKeys(ctx, key, favoriteMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
		} else {
			// 其他情况
			// 插入后，删除Redis中对应记录
			delErr := deleteKeys(ctx, key, favoriteMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
		}
	}
	return nil
}

func RelationMoveToDB() error {
	logger := zap.InitLogger()

	ctx := context.Background()
	keys, err := getKeys(ctx, "user::*::to_user::*::w")
	if err != nil {
		logger.Errorln(err)
		return err
	}
	for _, key := range keys {
		actionType, err := GetRedisHelper().Get(ctx, key).Result()
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		// 拆分得key
		kSplit := strings.Split(key, "::")
		uid, tid := kSplit[1], kSplit[3]
		userID, err := strconv.ParseInt(uid, 10, 64)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		toUserID, err := strconv.ParseInt(tid, 10, 64)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		// 查询是否存在关注记录
		relation, err := db.GetRelationByUserIDs(ctx, userID, toUserID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		} else if relation == nil && actionType == "1" {
			// 数据库中没有该关注记录，且最终状态为关注，则插入数据库
			err = db.CreateRelation(ctx, userID, toUserID)
			// 插入后，删除Redis中对应记录
			delErr := deleteKeys(ctx, key, relationMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
		} else if relation != nil && actionType == "2" {
			// 数据库中有该关注记录，且最终状态为取消关注，则从数据库中删除该记录
			err = db.DelRelationByUserIDs(ctx, userID, toUserID)
			// 删除Redis中对应记录
			delErr := deleteKeys(ctx, key, relationMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
		}
		// 删除Redis中对应记录
		delErr := deleteKeys(ctx, key, relationMutex)
		if delErr != nil {
			logger.Errorln(delErr.Error())
			return delErr
		}
	}
	return nil
}

func GoCronFavorite() {
	s := gocron.NewSchedule()
	s.Every(frequency).Tag("favoriteRedis").Seconds().Do(FavoriteMoveToDB)
	s.StartAsync()
}

func GoCronRelation() {
	s := gocron.NewSchedule()
	s.Every(frequency).Tag("relationRedis").Seconds().Do(RelationMoveToDB)
	s.StartAsync()
}
