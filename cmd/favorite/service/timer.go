package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/redis"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/gocron"
)

const frequency = 10

// 点赞服务消息队列消费者
func consume() error {
	msgs, err := FavoriteMq.ConsumeSimple()
	if err != nil {
		fmt.Println(err.Error())
		logger.Errorf("FavoriteMQ Err: %s", err.Error())
	}
	// 将消息队列的消息全部取出
	for msg := range msgs {
		//err := redis.LockByMutex(context.Background(), redis.FavoriteMutex)
		//if err != nil {
		//	logger.Errorf("Redis mutex lock error: %s", err.Error())
		//	return err
		//}
		fc := new(redis.FavoriteCache)
		// 解析json
		if err = json.Unmarshal(msg.Body, &fc); err != nil {
			logger.Errorf("json unmarshal error: %s", err.Error())
			fmt.Println("json unmarshal error:" + err.Error())
			//err = redis.UnlockByMutex(context.Background(), redis.FavoriteMutex)
			//if err != nil {
			//	logger.Errorf("Redis mutex unlock error: %s", err.Error())
			//	return err
			//}
			continue
		}
		fmt.Printf("==> Get new message: %v\n", fc)
		// 将结构体存入redis
		if err = redis.UpdateFavorite(context.Background(), fc); err != nil {
			logger.Errorf("json unmarshal error: %s", err.Error())
			fmt.Println("json unmarshal error:" + err.Error())
			//err = redis.UnlockByMutex(context.Background(), redis.FavoriteMutex)
			//if err != nil {
			//	logger.Errorf("Redis mutex unlock error: %s", err.Error())
			//	return err
			//}
			continue
		}
		//err = redis.UnlockByMutex(context.Background(), redis.FavoriteMutex)
		//if err != nil {
		//	logger.Errorf("Redis mutex unlock error: %s", err.Error())
		//	return err
		//}
		if !autoAck {
			err := msg.Ack(true)
			if err != nil {
				logger.Errorf("ack error: %s", err.Error())
				return err
			}
		}
	}
	return nil
}

// gocron定时任务,每隔一段时间就让Consumer消费消息队列的所有消息
func GoCron() {
	s := gocron.NewSchedule()
	s.Every(frequency).Tag("favoriteMQ").Seconds().Do(consume)
	s.StartAsync()
}
