package tool

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/dousheng/pkg/viper"
	"io/ioutil"
	"os"
)

// 提供各种加密算法
const (
	base64Table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

var (
	coder              = base64.NewEncoding(base64Table)
	config             = viper.Init("crypt")
	PublicKeyFilePath  = config.Viper.GetString("rsa.tiktok_message_encrypt_public_key")
	PrivateKeyFilePath = config.Viper.GetString("rsa.tiktok_message_decrypt_private_key")
)

func Base64Encode(src []byte) []byte {
	return []byte(coder.EncodeToString(src))
}

func Base64Decode(src []byte) ([]byte, error) {
	return coder.DecodeString(string(src))
}

func RsaEncrypt(origData []byte, publicKey string) ([]byte, error) {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		return nil, errors.New("public key error")
	}
	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	pub := pubInterface.(*rsa.PublicKey)
	return rsa.EncryptPKCS1v15(rand.Reader, pub, origData)
}

func RsaDecrypt(ciphertext []byte, privateKey string) ([]byte, error) {
	block, _ := pem.Decode([]byte(privateKey))
	if block == nil {
		return nil, errors.New("private key error")
	}
	priv, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	return rsa.DecryptPKCS1v15(rand.Reader, priv.(*rsa.PrivateKey), ciphertext)
}

func Sha256Encrypt(data string, salt string) string {
	sha256Ctx := sha256.New()            //sha256 init
	sha256Ctx.Write([]byte(data + salt)) //sha256 updata
	cipherStr := sha256.Sum256(nil)      //sha256 final
	return fmt.Sprintf("%x", cipherStr)
}

func Md5Encrypt(data string) string {
	md5Ctx := md5.New()                            //md5 init
	md5Ctx.Write([]byte(data))                     //md5 updata
	cipherStr := md5Ctx.Sum(nil)                   //md5 final
	encryptedData := hex.EncodeToString(cipherStr) //hex_digest
	return encryptedData
}

func ReadKeyFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return "", err
	}

	if err != nil {
		return "", err
	}

	return string(data), nil
}
