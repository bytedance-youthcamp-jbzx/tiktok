package service

import (
	"github.com/bytedance-youthcamp-jbzx/dousheng/pkg/jwt"
)

var (
	Jwt *jwt.JWT
)

func Init(signingKey string) {
	Jwt = jwt.NewJWT([]byte(signingKey))
}
