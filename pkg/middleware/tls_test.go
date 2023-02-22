package middleware

import (
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTLSServer(t *testing.T) {
	sync := make(chan int, 1)
	tlsKey := os.Getenv("tiktok_tls_key")
	tlsCert := os.Getenv("tiktok_tls_cert")

	if len(tlsCert) == 0 || len(tlsKey) == 0 {

		t.Fatalf("key or cert not found in environment")
	}

	go func(t *testing.T) {
		r := gin.Default()

		r.GET("/", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status_code": 1,
			})
		})

		r.Use(TLSSupportMiddleware("1.12.68.184:4001"))

		sync <- 1
		r.RunTLS("1.12.68.184:4001", tlsCert, tlsKey)
	}(t)

	<-sync
	time.Sleep(time.Second * 2)
	_, err := http.Get("https://1.12.68.184:4001/")

	if err != nil {
		t.Fatalf("https request error: %v", err)
	}
}
