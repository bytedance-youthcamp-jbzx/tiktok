package middleware

import (
	"context"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

// 限流中间件，使用令牌桶的方式处理请求。Note: auth中间件需在其前面
func TokenLimitMiddleware() app.HandlerFunc {
	logger := zap.InitLogger()

	return func(ctx context.Context, c *app.RequestContext) {
		token := c.GetString("Token")

		if !CurrentLimiter.Allow(token) {
			responseWithError(ctx, c, http.StatusForbidden, "request too fast")
			logger.Errorln("403: Request too fast.")
			return
		}
		c.Next(ctx)
	}
}
