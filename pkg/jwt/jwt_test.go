package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt"
)

func TestJWT(t *testing.T) {
	userJwt := NewJWT([]byte{0x12, 0x32, 0x4a, 0x53, 0x59, 0x45})

	token, err := userJwt.CreateToken(CustomClaims{
		1234,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 5).Unix(),
			Issuer:    "dousheng",
		},
	})

	if err != nil {
		t.Fatalf("create token error %v", err)
	}

	_, err = userJwt.ParseToken(token)

	if err != nil {
		t.Fatalf("token verified error %v", err)
	}

	otherJwt := NewJWT([]byte{0x12, 0x32, 0x4a, 0x53, 0x59, 0x45})
	_, err = otherJwt.ParseToken(token)

	if err != nil {
		t.Fatalf("token verified error %v", err)
	}

	time.Sleep(time.Second * 7)

	_, err = userJwt.ParseToken(token)

	if err == nil {
		t.Fatalf("token expired but not got error")
	}

}
