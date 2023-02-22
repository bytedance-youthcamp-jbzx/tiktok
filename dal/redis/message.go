package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

/*
轮询获取message需要redis记录前一次获取信息的最后一条消息的时间戳，键是用户的令牌，
值是上次消息的最后条消息的时间戳，规定键的过期时间为两秒。每次轮询的请求都需要去更新
redis里的键值，即便没有新的消息传来。
*/

func GetMessageTimestamp(ctx context.Context, token string, toUserID int64) (int, error) {
	key := fmt.Sprintf("%s_%d", token, toUserID)
	if ec, err := GetRedisHelper().Exists(ctx, key).Result(); err != nil {
		return 0, err
	} else if ec == 0 {
		return 0, nil //errors.New("key not found")
	}

	val, err := GetRedisHelper().Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(val)
}

func SetMessageTimestamp(ctx context.Context, token string, toUserID int64, timestamp int) error {
	key := fmt.Sprintf("%s_%d", token, toUserID)
	return GetRedisHelper().Set(ctx, key, timestamp, 2*time.Second).Err()
}
