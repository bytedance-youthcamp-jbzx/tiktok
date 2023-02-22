package service

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/tool"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
)

var (
	Jwt        *jwt.JWT
	publicKey  string
	privateKey string
)

func Init(signingKey string) {
	Jwt = jwt.NewJWT([]byte(signingKey))
	publicKey, _ = tool.ReadKeyFromFile(tool.PublicKeyFilePath)
	privateKey, _ = tool.ReadKeyFromFile(tool.PrivateKeyFilePath)
}
