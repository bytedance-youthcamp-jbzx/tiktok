package middleware

import (
	"context"
	z "github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/kitex/pkg/endpoint"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var (
	_      endpoint.Middleware = CommonMiddleware
	logger *zap.SugaredLogger
)

func init() {
	logger = z.InitLogger()
	defer logger.Sync()

	//klog.SetLogger(logger)
}

func responseWithError(ctx context.Context, c *app.RequestContext, code int, message interface{}) {
	c.AbortWithStatusJSON(code, gin.H{
		"status_code": -1, // 业务码 400x错误，建议细化
		"status_msg":  message,
	})
}

func CommonMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, req, resp interface{}) (err error) {
		ri := rpcinfo.GetRPCInfo(ctx)
		// get real request
		logger.Debugf("real request: %+v", req)
		// get remote service information
		logger.Debugf("remote service name: %s, remote method: %s", ri.To().ServiceName(), ri.To().Method())
		if err := next(ctx, req, resp); err != nil {
			return err
		}
		// get real response
		logger.Infof("real response: %+v", resp)
		return nil
	}
}
