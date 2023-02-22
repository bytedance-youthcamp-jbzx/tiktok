package rabbitmq

import (
	"fmt"

	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	z "github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	config = viper.Init("rabbitmq")
	logger *zap.SugaredLogger
	conn   *amqp.Connection
	err    error
	MqUrl  = fmt.Sprintf("amqp://%s:%s@%s:%d/%v",
		config.Viper.GetString("server.username"),
		config.Viper.GetString("server.password"),
		config.Viper.GetString("server.host"),
		config.Viper.GetInt("server.port"),
		config.Viper.GetString("server.vhost"),
	)
)

func init() {
	logger = z.InitLogger()
}

func failOnError(err error, msg string) {
	if err != nil {
		logger.Errorf("%s: %s", msg, err.Error())
		panic(fmt.Sprintf("%s: %s", msg, err.Error()))
	}
}
