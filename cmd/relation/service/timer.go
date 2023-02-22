package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance-youthcamp-jbzx/dousheng/dal/redis"
	"github.com/bytedance-youthcamp-jbzx/dousheng/pkg/gocron"
)

const frequency = 10

// 点赞服务消息队列消费者
func consume() {
	msgs, err := RelationMq.ConsumeSimple()
	if err != nil {
		fmt.Println(err.Error())
	}
	// 将消息队列的消息全部取出
	for msg := range msgs {
		fmt.Printf("==> Get new message: %v", msg.MessageId)
		rc := new(redis.RelationCache)
		// 解析json
		if err = json.Unmarshal(msg.Body, &rc); err != nil {
			fmt.Println("json unmarshal error:" + err.Error())
			continue
		}
		// 将结构体存入redis
		if err = redis.UpdateRelation(context.Background(), rc); err != nil {
			fmt.Println("add to redis error:" + err.Error())
			continue
		}
	}
}

// gocron定时任务,每隔一段时间就让Consumer消费消息队列的所有消息
func GoCron() {
	s := gocron.NewSchedule()
	s.Every(frequency).Tag("relationMQ").Seconds().Do(consume)
	s.StartAsync()
}
