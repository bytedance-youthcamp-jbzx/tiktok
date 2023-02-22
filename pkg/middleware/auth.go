package middleware

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func TokenAuthMiddleware(jwt jwt.JWT, skipRoutes ...string) gin.HandlerFunc {
	logger := zap.InitLogger()
	// TODO: signKey可以保存在环境变量中，而不是硬编码在代码里，可以通过获取环境变量的方式获得signkey
	return func(c *gin.Context) {
		// 对于skip的路由不对他进行token鉴权
		for _, skipRoute := range skipRoutes {
			if skipRoute == c.FullPath() {
				c.Next()
				return
			}
		}

		// 从处理get post请求中获取token
		var token string
		if c.Request.Method == "GET" {
			token = c.Query("token")
		} else if c.Request.Method == "POST" {
			if strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data") {
				token = c.PostForm("token")
			} else {
				token = c.Query("token")
			}
		} else {
			// Unsupport request method
			responseWithError(c, http.StatusBadRequest, "bad request")
			logger.Errorln("bad request")
			return
		}
		if token == "" {
			responseWithError(c, http.StatusUnauthorized, "token required")
			logger.Errorln("token required")
			// 提前返回
			return
		}

		claim, err := jwt.ParseToken(token)

		if err != nil {
			responseWithError(c, http.StatusUnauthorized, err.Error())
			logger.Errorln(err.Error())
			return
		}

		// 在上下文中向下游传递token
		c.Set("Token", token)
		c.Set("Id", claim.Id)

		c.Next() // 交给下游中间件
	}
}
