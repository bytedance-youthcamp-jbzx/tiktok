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
	msgs, err := RelationMq.ConsumeSimple()
	if err != nil {
		fmt.Println(err.Error())
		logger.Errorf("RelationMQ Err: %s", err.Error())
		return err
	}
	// 将消息队列的消息全部取出
	for msg := range msgs {
		//err := redis.LockByMutex(context.Background(), redis.RelationMutex)
		//if err != nil {
		//	logger.Errorf("Redis mutex lock error: %s", err.Error())
		//	return err
		//}
		rc := new(redis.RelationCache)
		// 解析json
		if err = json.Unmarshal(msg.Body, &rc); err != nil {
			fmt.Println("json unmarshal error:" + err.Error())
			logger.Errorf("RelationMQ Err: %s", err.Error())
			//err = redis.UnlockByMutex(context.Background(), redis.RelationMutex)
			//if err != nil {
			//	logger.Errorf("Redis mutex unlock error: %s", err.Error())
			//	return err
			//}
			continue
		}
		fmt.Printf("==> Get new message: %v\n", rc)
		// 将结构体存入redis
		if err = redis.UpdateRelation(context.Background(), rc); err != nil {
			fmt.Println("add to redis error:" + err.Error())
			logger.Errorf("RelationMQ Err: %s", err.Error())
			//err = redis.UnlockByMutex(context.Background(), redis.RelationMutex)
			//if err != nil {
			//	logger.Errorf("Redis mutex unlock error: %s", err.Error())
			//	return err
			//}
			continue
		}
		//err = redis.UnlockByMutex(context.Background(), redis.RelationMutex)
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
	s.Every(frequency).Tag("relationMQ").Seconds().Do(consume)
	s.StartAsync()
}
