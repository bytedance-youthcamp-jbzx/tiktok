package middleware

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	sjwt "github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"

	gjwt "github.com/golang-jwt/jwt"

	"github.com/gin-gonic/gin"
)

type authResponse struct {
	StatusCode int    `json:"status_code"`
	StatusMsg  string `json:"status_msg"`
}

type tokenResponse struct {
	authResponse
	Token string `json:"token"`
}

func runTestServer(middleware gin.HandlerFunc) {
	// 创建一个服务器包含创建token和token验证的服务器
	r := gin.Default()
	signKey := []byte{0x12, 0x34, 0x56, 0x78, 0x9a}
	userJwt := sjwt.NewJWT(signKey)

	r.Use(middleware)
	r.POST("/login", func(c *gin.Context) {
		username, err := strconv.Atoi(c.PostForm("username"))
		if err != nil {
			c.JSON(200, gin.H{
				"status_code": -1,
				"status_msg":  "invalid argument",
			})
			return
		}
		token, err := userJwt.CreateToken(sjwt.CustomClaims{
			int64(username),
			gjwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Second * 5).Unix(), // 5秒之后失效
				Issuer:    "dousheng",
			},
		})

		if err != nil {
			c.JSON(200, gin.H{
				"status_code": -1,
				"status_msg":  err,
			})

		} else {
			c.JSON(200, gin.H{
				"status_code": 0,
				"status_msg":  "",
				"token":       token,
			})
		}

	})

	r.GET("/user", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status_code": 0,
			"status_msg":  "",
		})

	})

	go func() {
		r.Run(":4001")
	}()

}

func doAuth(token string, t *testing.T) (int, int) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:4001/user?token=%s", token))

	if err != nil {
		t.Fatalf("error %v", err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatalf("error %v", err)
	}

	authResponse := authResponse{}

	err = json.Unmarshal(body, &authResponse)

	if err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	return resp.StatusCode, authResponse.StatusCode
}

func getAuthToken(username string, t *testing.T) string {
	resp, err := http.Post("http://localhost:4001/login",
		"application/x-www-form-urlencoded",
		strings.NewReader("username="+username))

	if err != nil {
		t.Fatalf("error %v", err)
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		t.Fatalf("error %v", err)
	}

	tokenResponse := tokenResponse{}

	err = json.Unmarshal(body, &tokenResponse)

	if err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	return tokenResponse.Token
}
