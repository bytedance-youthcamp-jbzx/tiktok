package service

import (
	"github.com/bytedance-youthcamp-jbzx/dousheng/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/dousheng/pkg/rabbitmq"
	"github.com/bytedance-youthcamp-jbzx/dousheng/pkg/zap"
)

var (
	Jwt        *jwt.JWT
	logger     = zap.InitLogger()
	FavoriteMq = rabbitmq.NewRabbitMQSimple("favorite")
	err        error
)

func Init(signingKey string) {
	Jwt = jwt.NewJWT([]byte(signingKey))
	GoCron()
}
