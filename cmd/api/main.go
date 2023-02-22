package main

import (
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/cmd/api/handler"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/middleware"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/gin-gonic/gin"
)

var (
	config        = viper.Init("api")
	apiServerAddr = fmt.Sprintf("%s:%d", config.Viper.GetString("server.host"), config.Viper.GetInt("server.port"))
	signingKey    = config.Viper.GetString("JWT.signingKey")
	enableTLS     = config.Viper.GetBool("tls.enable")
	serverTLSKey  string
	serverTLSCert string
)

func initGin() *gin.Engine {
	opts := []gin.HandlerFunc{
		middleware.TokenAuthMiddleware(*jwt.NewJWT([]byte(signingKey)),
			"/douyin/user/register/",
			"/douyin/user/login/",
			"/douyin/feed",
			"/douyin/favorite/list/",
			"/douyin/publish/list/",
			"/douyin/comment/list/",
			"/douyin/relation/follower/list/",
			"/douyin/relation/follow/list/",
		), // 用户鉴权中间件
		middleware.TokenLimitMiddleware(), //限流中间件
	}

	if enableTLS {
		serverTLSKey = config.Viper.GetString("tls.tiktok_tls_key")
		serverTLSCert = config.Viper.GetString("tls.tiktok_tls_cert")
		if len(serverTLSKey) == 0 {
			panic("not found tiktok_tls_key in configuration")
		}
		if len(serverTLSCert) == 0 {
			panic("not found tiktok_tls_cert in configuration")
		}
		opts = append(opts, middleware.TLSSupportMiddleware(apiServerAddr)) // TLS协议中间件
	}

	r := gin.Default()
	r.Use(opts...)
	return r
}

func registerGroup(r *gin.Engine) {
	douyin := r.Group("/douyin")
	{
		user := douyin.Group("/user")
		{
			user.GET("/", handler.UserInfo)
			user.POST("/register/", handler.Register)
			user.POST("/login/", handler.Login)
		}
		message := douyin.Group("/message")
		{
			message.GET("/chat/", handler.MessageChat)
			message.POST("/action/", handler.MessageAction)
		}
		relation := douyin.Group("/relation")
		{
			// 粉丝列表
			relation.GET("/follower/list/", handler.FollowerList)
			// 关注列表
			relation.GET("/follow/list/", handler.FollowList)
			// 朋友列表
			relation.GET("/friend/list/", handler.FriendList)
			relation.POST("/action/", handler.RelationAction)
		}
		publish := douyin.Group("/publish")
		{
			publish.GET("/list/", handler.PublishList)
			publish.POST("/action/", handler.PublishAction)
		}
		douyin.GET("/feed", handler.Feed)
		favorite := douyin.Group("/favorite")
		{
			favorite.POST("/action/", handler.FavoriteAction)
			favorite.GET("/list/", handler.FavoriteList)
		}
		comment := douyin.Group("/comment")
		{
			comment.POST("/action/", handler.CommentAction)
			comment.GET("/list/", handler.CommentList)
		}
	}
}

func main() {
	logger := zap.InitLogger()

	// initial gin
	r := initGin()
	// add handler
	registerGroup(r)

	if enableTLS {
		if err := r.RunTLS(apiServerAddr, serverTLSCert, serverTLSKey); err != nil {
			logger.Fatalln(err.Error())
		}
	} else {
		if err := r.Run(apiServerAddr); err != nil {
			logger.Fatalln(err.Error())
		}
	}
}
