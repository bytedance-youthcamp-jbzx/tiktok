package rpc

import "github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"

func init() {
	// comment rpc
	commentConfig := viper.Init("comment")
	InitComment(&commentConfig)

	// favorite rpc
	favoriteConfig := viper.Init("favorite")
	InitFavorite(&favoriteConfig)

	// message rpc
	messageConfig := viper.Init("message")
	InitMessage(&messageConfig)

	// relation rpc
	relationConfig := viper.Init("relation")
	InitRelation(&relationConfig)

	// user rpc
	userConfig := viper.Init("user")
	InitUser(&userConfig)

	// video rpc
	videoConfig := viper.Init("video")
	InitVideo(&videoConfig)
}
