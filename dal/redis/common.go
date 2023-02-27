package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/gocron"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/go-redsync/redsync/v4"
	"strconv"
	"strings"
	"time"
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

func setKey(ctx context.Context, key string, value string, expireTime time.Duration, mutex *redsync.Mutex) error {
	fmt.Println(key, " => ", value)
	_, err := GetRedisHelper().Set(ctx, key, value, expireTime).Result()
	errUnlock := UnlockByMutex(ctx, mutex)
	if errUnlock != nil {
		zapLogger.Errorf("unlock failed: %s", errUnlock.Error())

		return errUnlock
	}
	if err != nil {
		return errors.New("Redis set key failed: " + err.Error())
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
		LockByMutex(ctx, FavoriteMutex)
		res, err := GetRedisHelper().Get(ctx, key).Result()
		UnlockByMutex(ctx, FavoriteMutex)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		// 拆分得 value
		vSplit := strings.Split(res, "::")
		_, redisAt := vSplit[0], vSplit[1]
		// 拆分得 key
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

		// 检查是否存在对应ID
		v, err := db.GetVideoById(ctx, videoID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		u, err := db.GetUserByID(ctx, userID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		if v == nil || u == nil {
			delErr := deleteKeys(ctx, key, FavoriteMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			continue
		}

		// 查询是否存在点赞记录
		favorite, err := db.GetFavoriteVideoRelationByUserVideoID(ctx, userID, videoID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		} else if favorite == nil && redisAt == "1" {
			// 数据库中没有该点赞记录，且最终状态为点赞，则插入数据库
			video, err := db.GetVideoById(ctx, videoID)
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
			err = db.CreateVideoFavorite(ctx, userID, videoID, int64(video.AuthorID))
			// 插入后，删除Redis中对应记录
			delErr := deleteKeys(ctx, key, FavoriteMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
		} else if favorite != nil && redisAt == "2" {
			// 数据库中有该点赞记录，且最终状态为取消点赞，则从数据库中删除该记录
			video, err := db.GetVideoById(ctx, videoID)
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
			err = db.DelFavoriteByUserVideoID(ctx, userID, videoID, int64(video.AuthorID))
			// 插入后，删除Redis中对应记录
			delErr := deleteKeys(ctx, key, FavoriteMutex)
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
			delErr := deleteKeys(ctx, key, FavoriteMutex)
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
		res, err := GetRedisHelper().Get(ctx, key).Result()
		vSplit := strings.Split(res, "::")
		_, redisAt := vSplit[0], vSplit[1]
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

		// 检查是否存在对应ID
		u, err := db.GetUserByID(ctx, userID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		tu, err := db.GetUserByID(ctx, toUserID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		}
		if u == nil || tu == nil {
			delErr := deleteKeys(ctx, key, RelationMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			continue
		}

		// 查询是否存在关注记录
		relation, err := db.GetRelationByUserIDs(ctx, userID, toUserID)
		if err != nil {
			logger.Errorln(err.Error())
			return err
		} else if relation == nil && redisAt == "1" {
			// 数据库中没有该关注记录，且最终状态为关注，则插入数据库
			err = db.CreateRelation(ctx, userID, toUserID)
			// 插入后，删除Redis中对应记录
			delErr := deleteKeys(ctx, key, RelationMutex)
			if delErr != nil {
				logger.Errorln(delErr.Error())
				return delErr
			}
			if err != nil {
				logger.Errorln(err.Error())
				return err
			}
		} else if relation != nil && redisAt == "2" {
			// 数据库中有该关注记录，且最终状态为取消关注，则从数据库中删除该记录
			err = db.DelRelationByUserIDs(ctx, userID, toUserID)
			// 删除Redis中对应记录
			delErr := deleteKeys(ctx, key, RelationMutex)
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
		delErr := deleteKeys(ctx, key, RelationMutex)
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
