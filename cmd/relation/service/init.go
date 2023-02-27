// Package service /*
package service

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/tool"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/rabbitmq"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
)

var (
	Jwt        *jwt.JWT
	logger     = zap.InitLogger()
	config     = viper.Init("rabbitmq")
	autoAck    = config.Viper.GetBool("consumer.relation.autoAck")
	RelationMq = rabbitmq.NewRabbitMQSimple("relation", autoAck)
	err        error
	privateKey string
)

func Init(signingKey string) {
	Jwt = jwt.NewJWT([]byte(signingKey))
	privateKey, _ = tool.ReadKeyFromFile(tool.PrivateKeyFilePath)
	//GoCron()
	go consume()
}
