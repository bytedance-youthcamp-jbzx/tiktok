package middleware

import (
	"sync"
	"sync/atomic"
	"testing"
)

func TestLimit(t *testing.T) {
	// 启动测试服务器，安装限流中间件
	runTestServer(TokenLimitMiddleware())

	token := getAuthToken("123", t)
	nums := 20
	var forbidden int32
	var worker sync.WaitGroup
	worker.Add(nums)

	for i := 0; i < nums; i++ {
		go func(t *testing.T) {
			code, _ := doAuth(token, t)
			if code == 403 {
				atomic.AddInt32(&forbidden, 1)
			}
			worker.Done()
		}(t)
	}

	worker.Wait()

	if forbidden == 0 {
		t.Fatalf("forbidden must > 0")
	}

}
