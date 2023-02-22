package middleware

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/gin-gonic/gin"
)

// 限流中间件，使用令牌桶的方式处理请求。Note: auth中间件需在其前面
func TokenLimitMiddleware() gin.HandlerFunc {
	logger := zap.InitLogger()

	return func(c *gin.Context) {
		token := c.GetString("Token")

		if !CurrentLimiter.Allow(token) {
			responseWithError(c, 403, "request too fast")
			logger.Errorln("403: Request too fast.")
			return
		}
		c.Next()
	}
}
