package service

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/rabbitmq"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
)

var (
	Jwt        *jwt.JWT
	logger     = zap.InitLogger()
	config     = viper.Init("rabbitmq")
	autoAck    = config.Viper.GetBool("consumer.favorite.autoAck")
	FavoriteMq = rabbitmq.NewRabbitMQSimple("favorite", autoAck)
	err        error
)

func Init(signingKey string) {
	Jwt = jwt.NewJWT([]byte(signingKey))
	//GoCron()
	go consume()
}
