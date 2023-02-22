// Package rpc /*
package rpc

import (
	"context"
	"fmt"
	"time"

	message "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/message"
	"github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/message/messageservice"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/etcd"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/middleware"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/retry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

var (
	messageClient messageservice.Client
)

func InitMessage(config *viper.Config) {
	etcdAddr := fmt.Sprintf("%s:%d", config.Viper.GetString("etcd.host"), config.Viper.GetInt("etcd.port"))
	serviceName := config.Viper.GetString("server.name")
	r, err := etcd.NewEtcdResolver([]string{etcdAddr})
	if err != nil {
		panic(err)
	}

	c, err := messageservice.NewClient(
		serviceName,
		client.WithMiddleware(middleware.CommonMiddleware),
		client.WithInstanceMW(middleware.ClientMiddleware),
		client.WithMuxConnection(1),                       // mux
		client.WithRPCTimeout(30*time.Second),             // rpc timeout
		client.WithConnectTimeout(30000*time.Millisecond), // conn timeout
		client.WithFailureRetry(retry.NewFailurePolicy()), // retry
		//client.WithSuite(tracing.NewClientSuite()),        // tracer
		client.WithResolver(r), // resolver
		// Please keep the same as provider.WithServiceName
		client.WithClientBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
	)
	if err != nil {
		panic(err)
	}
	messageClient = c
}

func MessageAction(ctx context.Context, req *message.MessageActionRequest) (*message.MessageActionResponse, error) {
	return messageClient.MessageAction(ctx, req)
}

func MessageChat(ctx context.Context, req *message.MessageChatRequest) (*message.MessageChatResponse, error) {
	return messageClient.MessageChat(ctx, req)
}
