package middleware

import (
	"context"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"strings"
)

func TokenAuthMiddleware(jwt jwt.JWT, skipRoutes ...string) app.HandlerFunc {
	logger := zap.InitLogger()
	// TODO: signKey可以保存在环境变量中，而不是硬编码在代码里，可以通过获取环境变量的方式获得signkey
	return func(ctx context.Context, c *app.RequestContext) {
		// 对于skip的路由不对他进行token鉴权
		for _, skipRoute := range skipRoutes {
			if skipRoute == c.FullPath() {
				c.Next(ctx)
				return
			}
		}

		// 从处理get post请求中获取token
		var token string
		if string(c.Request.Method()[:]) == "GET" {
			token = c.Query("token")
		} else if string(c.Request.Method()[:]) == "POST" {
			if strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
				token = c.PostForm("token")
			} else {
				token = c.Query("token")
			}
		} else {
			// Unsupport request method
			responseWithError(ctx, c, http.StatusBadRequest, "bad request")
			logger.Errorln("bad request")
			return
		}
		if token == "" {
			responseWithError(ctx, c, http.StatusUnauthorized, "token required")
			logger.Errorln("token required")
			// 提前返回
			return
		}

		claim, err := jwt.ParseToken(token)

		if err != nil {
			responseWithError(ctx, c, http.StatusUnauthorized, err.Error())
			logger.Errorln(err.Error())
			return
		}

		// 在上下文中向下游传递token
		c.Set("Token", token)
		c.Set("Id", claim.Id)

		c.Next(ctx) // 交给下游中间件
	}
}
