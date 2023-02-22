package service

import (
	"bytes"
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/tool"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/minio"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
)

// uploadVideo 上传视频至 Minio
func uploadVideo(data []byte, videoTitle string) (string, error) {
	logger := zap.InitLogger()

	// 将视频数据上传至minio
	reader := bytes.NewReader(data)
	contentType := "application/mp4"

	uploadSize, err := minio.UploadFileByIO(minio.VideoBucketName, videoTitle, reader, int64(len(data)), contentType)
	if err != nil {
		logger.Errorf("视频上传至minio失败：%v", err.Error())
		return "", err
	}
	logger.Infof("视频文件大小为：%v", uploadSize)
	fmt.Println("视频文件大小为：", uploadSize)

	// 获取上传文件的路径
	playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, videoTitle)
	if err != nil {
		logger.Errorf("服务器内部错误：视频获取失败：%s", err.Error())
		return "", err
	}
	logger.Infof("上传视频路径：%v", playUrl)
	return playUrl, nil
}

// uploadCover 截取并上传封面至 Minio
func uploadCover(playUrl string, coverTitle string) error {
	logger := zap.InitLogger()

	// 截取第一帧并将图像上传至minio
	imgBuffer, err := tool.GetSnapshotImageBuffer(playUrl, 1)
	if err != nil {
		logger.Errorf("服务器内部错误，封面获取失败：%s", err.Error())
		return err
	}
	var imgByte []byte
	imgBuffer.Write(imgByte)
	contentType := "image/png"

	uploadSize, err := minio.UploadFileByIO(minio.CoverBucketName, coverTitle, imgBuffer, int64(imgBuffer.Len()), contentType)
	if err != nil {
		logger.Errorf("封面上传至minio失败：%v", err.Error())
		return err
	}

	// 获取上传文件的路径
	coverUrl, err := minio.GetFileTemporaryURL(minio.CoverBucketName, coverTitle)
	if err != nil {
		logger.Errorf("封面获取链接失败：%v", err.Error())
		return err
	}
	logger.Infof("上传封面路径：%v", coverUrl)
	logger.Infof("封面文件大小为：%v", uploadSize)
	fmt.Println("封面文件大小为：", uploadSize)

	return nil
}

// VideoPublish 上传视频并获取封面
func VideoPublish(data []byte, videoTitle string, coverTitle string) error {
	playUrl, err := uploadVideo(data, videoTitle)
	err = uploadCover(playUrl, coverTitle)
	if err != nil {
		return err
	}
	return err
}

func handleVideoPublish() {

}
