package tool

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func TestSha256Encrypt(t *testing.T) {
	plaintext := "helloworld"
	salt := os.Getenv("tiktok_password_sha256_salt")

	if len(salt) == 0 {
		t.Fatalf("salt not found in environment")
	}

	fmt.Println(Sha256Encrypt(plaintext, salt))
}

func readKeyFromFile(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(file)

	if err != nil {
		return "", err
	}

	return string(data), nil
}
func TestRsaCrypt(t *testing.T) {
	publicKeyFilePath := PublicKeyFilePath
	privateKeyFilePath := PrivateKeyFilePath

	if len(publicKeyFilePath) == 0 || len(privateKeyFilePath) == 0 {
		t.Fatalf("key path not found in environment")
	}

	publicKey, err := readKeyFromFile(publicKeyFilePath)

	if err != nil {
		t.Fatalf("read public key error: %v", err)
	}

	privateKey, err := readKeyFromFile(privateKeyFilePath)

	if err != nil {
		t.Fatalf("read private key error: %v", err)
	}

	plaintext := "helloworld"

	ciphertext, err := RsaEncrypt([]byte(plaintext), publicKey)

	if err != nil {
		t.Fatalf("rsa encrypt error: %v", err)
	}

	decrypted, err := RsaDecrypt(ciphertext, privateKey)

	if err != nil {
		t.Fatalf("rsa decrypt error: %v", err)
	}

	if string(decrypted) != plaintext {
		t.Fatal("decrypted text is inconsistent with the original")
	}

}
