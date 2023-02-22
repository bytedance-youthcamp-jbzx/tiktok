package service

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
)

var (
	Jwt *jwt.JWT
)

func Init(signingKey string) {
	Jwt = jwt.NewJWT([]byte(signingKey))
}
