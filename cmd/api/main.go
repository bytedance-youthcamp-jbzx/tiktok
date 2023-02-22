package main

import (
	"crypto/tls"
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/cmd/api/handler"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/middleware"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	z "github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network/standard"
	"github.com/hertz-contrib/gzip"
)

var (
	apiConfig     = viper.Init("api")
	apiServerName = apiConfig.Viper.GetString("server.name")
	apiServerAddr = fmt.Sprintf("%s:%d", apiConfig.Viper.GetString("server.host"), apiConfig.Viper.GetInt("server.port"))
	etcdAddress   = fmt.Sprintf("%s:%d", apiConfig.Viper.GetString("Etcd.host"), apiConfig.Viper.GetInt("Etcd.port"))
	signingKey    = apiConfig.Viper.GetString("JWT.signingKey")
	serverTLSKey  = apiConfig.Viper.GetString("Hertz.tls.keyFile")
	serverTLSCert = apiConfig.Viper.GetString("Hertz.tls.certFile")
)

func registerGroup(hz *server.Hertz) {
	douyin := hz.Group("/douyin")
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

func InitHertz() *server.Hertz {
	logger := z.InitLogger()

	opts := []config.Option{server.WithHostPorts(apiServerAddr)}

	// 服务注册
	//if apiConfig.Viper.GetBool("Etcd.enable") {
	//	r, err := etcd.NewEtcdRegistry([]string{etcdAddress})
	//	if err != nil {
	//		logger.Fatalln(err.Error())
	//	}
	//	opts = append(opts, server.WithRegistry(r, &registry.Info{
	//		ServiceName: apiServerName,
	//		Addr:        utils.NewNetAddr("tcp", apiServerAddr),
	//		Weight:      10,
	//		Tags:        nil,
	//	}))
	//}

	// 网络库
	hertzNet := standard.NewTransporter
	//if apiConfig.Viper.GetBool("Hertz.useNetPoll") {
	//	hertzNet = netpoll.NewTransporter
	//}
	opts = append(opts, server.WithTransport(hertzNet))

	// TLS & Http2
	// https://github.com/cloudwego/hertz-examples/blob/main/protocol/tls/main.go
	tlsConfig := tls.Config{
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		},
	}
	if apiConfig.Viper.GetBool("Hertz.tls.enable") {
		if len(serverTLSKey) == 0 {
			panic("not found tiktok_tls_key in configuration")
		}
		if len(serverTLSCert) == 0 {
			panic("not found tiktok_tls_cert in configuration")
		}

		cert, err := tls.LoadX509KeyPair(serverTLSCert, serverTLSKey)
		if err != nil {
			logger.Errorln(err)
		}
		tlsConfig.Certificates = append(tlsConfig.Certificates, cert)
		opts = append(opts, server.WithTLS(&tlsConfig))

		if alpn := apiConfig.Viper.GetBool("Hertz.tls.ALPN"); alpn {
			opts = append(opts, server.WithALPN(alpn))
		}
	} else if apiConfig.Viper.GetBool("Hertz.http2.enable") {
		opts = append(opts, server.WithH2C(apiConfig.Viper.GetBool("Hertz.http2.enable")))
	}

	hz := server.Default(opts...)

	hz.Use(
		// secure.New(
		// 	secure.WithSSLHost(apiServerAddr),
		// 	secure.WithSSLRedirect(true),
		// ),	// TLS
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
		middleware.AccessLog(),
		gzip.Gzip(gzip.DefaultCompression),
	)
	return hz
}

func main() {
	hz := InitHertz()

	// add handler
	registerGroup(hz)

	hz.Spin()
}
