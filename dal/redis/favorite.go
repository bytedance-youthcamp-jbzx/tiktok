package redis

import (
	"context"
	"fmt"
	"strconv"
)

type FavoriteCache struct {
	VideoID    uint `json:"video_id" redis:"video_id"`
	UserID     uint `json:"user_id" redis:"user_id"`
	ActionType uint `json:"action_type" redis:"action_type"` // 若redis中vid-uid相应的action_type=2，则表示取消点赞，不插入数据库
	//CreatedAt  time.Time `json:"created_at" redis:"created_at"`
}

/*
* UpdateFavorite
* ActionType == 1 点赞；ActionType == 2 取消点赞
* 1. 使用集合set来存储被点赞类型的id，Key为video+被点赞视频的id，Value为点赞的用户id列表
* 2. set存储某个类型点赞的记录，Key为video::vid::user::uid，hashKey为点赞视频+点赞人，Value为action_type
 */
func UpdateFavorite(ctx context.Context, favorite *FavoriteCache) error {
	//keyVideo := fmt.Sprintf("video::%d", favorite.VideoID)
	// Read 用于与前端同步，且创建定时器检查是否过期；Write 用于与前端同步，不设置过期，但是需要定时与MySQL同步后进行删除
	keyUserIDRead := fmt.Sprintf("video::%d::user::%d::r", favorite.VideoID, favorite.UserID)
	keyUserIDWrite := fmt.Sprintf("video::%d::user::%d::w", favorite.VideoID, favorite.UserID)
	//if favorite.ActionType == 1 {
	//	_, err := GetRedisHelper().SAdd(ctx, keyVideo, favorite.UserID).Result()
	//	if err != nil {
	//		zapLogger.Errorln(err.Error())
	//		return err
	//	}
	//	//userMap := make(map[string]interface{})
	//	//userMap["created_at"] = favorite.CreatedAt.Format("2006-01-02 15:04:05")
	//	//_, err = GetRedisHelper().Set(ctx, keyUserID, favorite.CreatedAt.Format("2006-01-02 15:04:05"), 0).Result()
	//} else if favorite.ActionType == 2 {
	//	err := GetRedisHelper().SRem(ctx, keyVideo, 1, keyUserIDRead).Err()
	//	if err != nil {
	//		zapLogger.Errorln(err.Error())
	//		return err
	//	}
	//} else {
	//	zapLogger.Errorln("\"action_type\" is not equal to 1 or 2")
	//	return errors.New("\"action_type\" is not equal to 1 or 2")
	//}
	_, err := GetRedisHelper().Set(ctx, keyUserIDRead, favorite.ActionType, ExpireTime).Result()
	if err != nil {
		zapLogger.Errorln(err.Error())
		return err
	}
	err = LockByMutex(ctx, favoriteMutex)
	if err != nil {
		zapLogger.Errorf("lock failed: %s", err.Error())
		return err
	}
	_, err1 := GetRedisHelper().Set(ctx, keyUserIDWrite, favorite.ActionType, 0).Result()
	err = UnlockByMutex(ctx, favoriteMutex)
	if err != nil {
		zapLogger.Errorf("unlock failed: %s", err.Error())
		return err
	}
	if err1 != nil {
		zapLogger.Errorln(err1.Error())
		return err1
	}
	return nil
}

/**
 * GetAllUserIDs
 * 获取所有被点赞类型id的点赞用户id
 * video::<video_id> -> user_id 列表
 */
func GetAllUserIDs(ctx context.Context, videoID int64) ([]int64, error) {
	key := fmt.Sprintf("video::%d", videoID)
	results, err := GetRedisHelper().SMembers(ctx, key).Result()
	if err != nil {
		zapLogger.Errorln(err.Error())
		return nil, err
	}
	userIDs := make([]int64, 0)
	for _, result := range results {
		id, _ := strconv.ParseInt(result, 10, 64)
		userIDs = append(userIDs, id)
	}
	return userIDs, nil
}

/**
 * GetUsersFavorites
 * 根据多个hashKey获取对应值
 * video::<video_id>::user::<user_id> -> FavoriteCache
 */
func GetUsersFavorites(ctx context.Context, videoID int64, userIDs []int64) ([]*FavoriteCache, error) {
	favorites := make([]*FavoriteCache, 0)
	for _, userID := range userIDs {
		favoriteCache, err := GetRedisHelper().Get(ctx, fmt.Sprintf("video::%d::user::%d::r", videoID, userID)).Result()
		if err != nil {
			zapLogger.Errorln(err.Error())
			return nil, err
		}
		//createdAt, err := time.ParseInLocation("2006-01-02 15:04:05", favoriteCache, time.Local)
		actionType, err := strconv.ParseInt(favoriteCache, 10, 64)
		if err != nil {
			zapLogger.Errorln(err.Error())
			return nil, err
		}
		favorites = append(favorites, &FavoriteCache{
			VideoID:    uint(videoID),
			UserID:     uint(userID),
			ActionType: uint(actionType),
		})
	}
	return favorites, nil
}
