package minio

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	minioClient               *minio.Client
	minioConfig               = viper.Init("minio")
	MinioEndPoint             = minioConfig.Viper.GetString("minio.Endpoint")
	MinioAccessKeyId          = minioConfig.Viper.GetString("minio.AccessKeyId")
	MinioSecretAccessKey      = minioConfig.Viper.GetString("minio.SecretAccessKey")
	UseSSL                    = minioConfig.Viper.GetBool("minio.UseSSL")
	VideoBucketName           = minioConfig.Viper.GetString("minio.VideoBucketName")
	CoverBucketName           = minioConfig.Viper.GetString("minio.CoverBucketName")
	AvatarBucketName          = minioConfig.Viper.GetString("minio.AvatarBucketName")
	BackgroundImageBucketName = minioConfig.Viper.GetString("minio.BackgroundImageBucketName")
	ExpireTime                = minioConfig.Viper.GetUint32("minio.ExpireTime")
)

func init() {
	s3client, err := minio.New(MinioEndPoint, &minio.Options{
		Creds:  credentials.NewStaticV4(MinioAccessKeyId, MinioSecretAccessKey, ""),
		Secure: UseSSL,
	})

	if err != nil {
		panic(err)
	}

	minioClient = s3client

	if err := CreateBucket(VideoBucketName); err != nil {
		panic(err)
	}
}
