package middleware

import (
	"testing"
	"time"

	sjwt "github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
)

func TestAuth(t *testing.T) {
	// 启动测试服务器，安装token鉴权中间件
	signKey := []byte{0x12, 0x34, 0x56, 0x78, 0x9a}
	userJwt := sjwt.NewJWT(signKey)
	runTestServer(TokenAuthMiddleware(*userJwt, "/login"))
	// 不存在的token
	token := "aaabbcccdd"
	_, statusCode := doAuth(token, t)
	if statusCode != -1 {
		t.Fatalf("expected %d but got %d", -1, statusCode)
	}

	// 在有效期里的token
	token = getAuthToken("123", t)
	_, statusCode = doAuth(token, t)
	if statusCode != 0 {
		t.Fatalf("expected %d but got %d", 0, statusCode)
	}

	// token过期
	time.Sleep(8 * time.Second)
	_, statusCode = doAuth(token, t)
	if statusCode != -1 {
		t.Fatalf("expected %d but got %d", -1, statusCode)
	}

}
