package minio

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"runtime/debug"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func ExpectEqual(left interface{}, right interface{}, t *testing.T) {
	if left != right {
		t.Fatalf("expected %v == %v\n%s", left, right, debug.Stack())
	}
}

func ExpectUnEqual(left interface{}, right interface{}, t *testing.T) {
	if left == right {
		t.Fatalf("expected %v != %v\n%s", left, right, debug.Stack())
	}
}

func TestCreateInvalidBucketName(t *testing.T) {
	// 创建非法桶
	err := CreateBucket("")
	ExpectUnEqual(err, nil, t)
}

func TestCreateValidBucketName(t *testing.T) {
	// 创建合法桶
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		minioClient.RemoveBucket(ctx, tmpBucketName)
	}
	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)

	exist, errEx = minioClient.BucketExists(ctx, tmpBucketName)

	ExpectEqual(exist, true, t)
	ExpectEqual(errEx, nil, t)

	minioClient.RemoveBucket(ctx, tmpBucketName)
}

func TestUploadFileByPathWhenFileNotExist(t *testing.T) {
	// 测试上传不存在的文件
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	tmpObjectName := "abcdefg"
	tmpFilePath := "filenotexist.mp4"
	contentType := "application/mp4"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		minioClient.RemoveBucket(ctx, tmpBucketName)
	}

	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)

	size, err := UploadFileByPath(tmpBucketName, tmpObjectName, tmpFilePath, contentType)

	ExpectEqual(size, int64(-1), t)
	ExpectUnEqual(err, nil, t)

	minioClient.RemoveBucket(ctx, tmpBucketName)
}

func TestUploadFileByPathWhenFileExist(t *testing.T) {
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	tmpObjectName := "abcdefg"
	tmpFilePath := "fileexist.mp4"
	contentType := "application/mp4"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		minioClient.RemoveBucket(ctx, tmpBucketName)
	}
	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)

	// 检查文件是否存在并获取其大小
	file, err := os.Open(tmpFilePath)
	ExpectEqual(err, nil, t)
	defer file.Close()

	fileStat, err := file.Stat()
	ExpectEqual(err, nil, t)

	size, err := UploadFileByPath(tmpBucketName, tmpObjectName, tmpFilePath, contentType)

	ExpectEqual(size, fileStat.Size(), t)
	ExpectEqual(err, nil, t)

	minioClient.RemoveBucket(ctx, tmpBucketName)
}

func TestGetFileURLWhenObjNotExist(t *testing.T) {
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	tmpObjectName := "aasdjals923ijnsjnfao3i"
	contentType := "application/mp4"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		minioClient.RemoveBucket(ctx, tmpBucketName)
	}
	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)

	url, _ := GetFileTemporaryURL(tmpBucketName, tmpObjectName)

	resp, err := http.Get(url)
	ExpectEqual(err, nil, t)
	ExpectUnEqual(resp.Header.Get("Content-Type"), contentType, t)

	minioClient.RemoveBucket(ctx, tmpBucketName)
}

func TestGetFileURLWhenObjExist(t *testing.T) {
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	tmpObjectName := "aasdjals923ijnsjnfao3i"
	tmpFilePath := "fileexist.mp4"
	contentType := "application/mp4"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		minioClient.RemoveBucket(ctx, tmpBucketName)
	}
	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)

	_, err = UploadFileByPath(tmpBucketName, tmpObjectName, tmpFilePath, contentType)
	ExpectEqual(err, nil, t)

	url, _ := GetFileTemporaryURL(tmpBucketName, tmpObjectName)
	resp, err := http.Get(url)

	ExpectEqual(err, nil, t)
	ExpectEqual(resp.Header.Get("Content-Type"), contentType, t)

	minioClient.RemoveBucket(ctx, tmpBucketName)
}

func TestGetFileURLWhenObjExpire(t *testing.T) {
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	tmpObjectName := "aasdjals923ijnsjnfao3i"
	tmpFilePath := "fileexist.mp4"
	contentType := "application/mp4"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		minioClient.RemoveBucket(ctx, tmpBucketName)
	}
	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)

	_, err = UploadFileByPath(tmpBucketName, tmpObjectName, tmpFilePath, contentType)
	ExpectEqual(err, nil, t)

	ExpireTime = 1

	url, _ := GetFileTemporaryURL(tmpBucketName, tmpObjectName)
	resp, err := http.Get(url)

	ExpectEqual(err, nil, t)
	ExpectEqual(resp.Header.Get("Content-Type"), contentType, t)

	time.Sleep(time.Second)
	resp, _ = http.Get(url)

	ExpectEqual(resp.StatusCode, 403, t)

	minioClient.RemoveBucket(ctx, tmpBucketName)
}

func TestUploadFileByIO(t *testing.T) {
	// 测试使用IO的方式上传文件到对象存储
	tmpBucketName := "dc4a65747c0646c4a6eed59fb5404617"
	tmpObjectName := "aasdjals923ijnsjnfao3i"
	tmpFilePath := "fileexist.mp4"
	contentType := "application/mp4"
	ctx := context.Background()
	exist, errEx := minioClient.BucketExists(ctx, tmpBucketName)
	if exist && errEx != nil {
		minioClient.RemoveBucket(ctx, tmpBucketName)
	}
	err := CreateBucket(tmpBucketName)
	ExpectEqual(err, nil, t)
	// 1. 搭建一个最简单的web服务器用于接收文件
	r := gin.Default()
	r.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		ExpectEqual(err, nil, t)
		fp, err := file.Open()
		ExpectEqual(err, nil, t)
		size, err := UploadFileByIO(tmpBucketName, tmpObjectName, fp, file.Size, contentType)
		ExpectEqual(size, file.Size, t)
		ExpectEqual(err, nil, t)
		minioClient.RemoveBucket(ctx, tmpBucketName)
	})
	go func() {
		r.Run("127.0.0.1:4001")
	}()
	// 2. 使用http client上传视频

	time.Sleep(5 * time.Second)
	file, err := os.Open(tmpFilePath)
	ExpectEqual(err, nil, t)
	defer file.Close()

	content := &bytes.Buffer{}
	writer := multipart.NewWriter(content)
	form, err := writer.CreateFormFile("file", filepath.Base(tmpFilePath))
	ExpectEqual(err, nil, t)
	_, err = io.Copy(form, file)
	ExpectEqual(err, nil, t)
	err = writer.Close()
	ExpectEqual(err, nil, t)
	req, err := http.NewRequest("POST", "http://127.0.0.1:4001/upload", content)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	ExpectEqual(err, nil, t)
	client := &http.Client{}
	_, err = client.Do(req)
	ExpectEqual(err, nil, t)

}

func TestGetFileTemporaryURL(t *testing.T) {
	tmpBucketName := "tiktok-videos"
	tmpObjectName := "7196183298031488288.mp4"
	url, err := GetFileTemporaryURL(tmpBucketName, tmpObjectName)
	//ExpectEqual(err, nil, t)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(url)
}
