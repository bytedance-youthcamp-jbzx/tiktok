package viper

import (
	"log"

	V "github.com/spf13/viper"
)

// Config 公有变量,获取Viper
type Config struct {
	Viper *V.Viper
}

// Init 初始化Viper配置
func Init(configName string) Config {
	config := Config{Viper: V.New()}
	v := config.Viper
	v.SetConfigType("yml")      //设置配置文件类型
	v.SetConfigName(configName) //设置配置文件名
	v.AddConfigPath("./config") //设置配置文件路径 !!!注意路径问题
	v.AddConfigPath("../config")
	v.AddConfigPath("../../config")
	//读取配置文件
	if err := v.ReadInConfig(); err != nil {
		//global.SugarLogger.Fatalf("read config files failed,errors is %+v", err)
		log.Fatalf("errno is %+v", err)
	}
	return config
}
