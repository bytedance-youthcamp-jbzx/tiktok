package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/tool"
	user "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
	video "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/video"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/minio"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/viper"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
)

// VideoServiceImpl implements the last service interface defined in the IDL.
type VideoServiceImpl struct{}

const limit = 30

// Feed implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) Feed(ctx context.Context, req *video.FeedRequest) (resp *video.FeedResponse, err error) {
	logger := zap.InitLogger()
	nextTime := time.Now().UnixMilli()
	var userID int64 = -1

	// 验证token有效性
	if req.Token != "" {
		claims, err := Jwt.ParseToken(req.Token)
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "token 解析错误",
			}
			return res, nil
		}
		userID = claims.Id
	}
	// 调用数据库查询 video_list
	videos, err := db.MGetVideos(ctx, limit, &req.LatestTime)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.FeedResponse{
			StatusCode: -1,
			StatusMsg:  "视频获取失败：服务器内部错误",
		}
		return res, nil
	}
	videoList := make([]*video.Video, 0)
	for _, r := range videos {
		author, err := db.GetUserByID(ctx, int64(r.AuthorID))
		if err != nil {
			logger.Errorf("error:%v", err.Error())
			return nil, err
		}
		relation, err := db.GetRelationByUserIDs(ctx, userID, int64(author.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "视频获取失败：服务器内部错误",
			}
			return res, nil
		}
		favorite, err := db.GetFavoriteVideoRelationByUserVideoID(ctx, userID, int64(r.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "视频获取失败：服务器内部错误",
			}
			return res, nil
		}
		playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, r.PlayUrl)
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：视频获取失败",
			}
			return res, nil
		}
		coverUrl, err := minio.GetFileTemporaryURL(minio.CoverBucketName, r.CoverUrl)
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：封面获取失败",
			}
			return res, nil
		}
		avatarUrl, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, author.Avatar)
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：头像获取失败",
			}
			return res, nil
		}
		backgroundUrl, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, author.BackgroundImage)
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &video.FeedResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：背景图获取失败",
			}
			return res, nil
		}

		videoList = append(videoList, &video.Video{
			Id: int64(r.ID),
			Author: &user.User{
				Id:              int64(author.ID),
				Name:            author.UserName,
				FollowCount:     int64(author.FollowingCount),
				FollowerCount:   int64(author.FollowerCount),
				IsFollow:        relation != nil,
				Avatar:          avatarUrl,
				BackgroundImage: backgroundUrl,
				Signature:       author.Signature,
				TotalFavorited:  int64(author.TotalFavorited),
				WorkCount:       int64(author.WorkCount),
				FavoriteCount:   int64(author.FavoriteCount),
			},
			PlayUrl:       playUrl,
			CoverUrl:      coverUrl,
			FavoriteCount: int64(r.FavoriteCount),
			CommentCount:  int64(r.CommentCount),
			IsFavorite:    favorite != nil,
			Title:         r.Title,
		})
	}
	if len(videos) != 0 {
		nextTime = videos[len(videos)-1].UpdatedAt.UnixMilli()
	}
	res := &video.FeedResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		VideoList:  videoList,
		NextTime:   nextTime,
	}
	return res, nil
}

// PublishAction implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) PublishAction(ctx context.Context, req *video.PublishActionRequest) (resp *video.PublishActionResponse, err error) {
	logger := zap.InitLogger()
	// 解析token,获取用户id
	claims, err := Jwt.ParseToken(req.Token)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "token 解析错误",
		}
		return res, nil
	}
	userID := claims.Id

	if len(req.Title) == 0 || len(req.Title) > 32 {
		logger.Errorf("标题不能为空且不能超过32个字符：%d", len(req.Title))
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "标题不能为空且不能超过32个字符",
		}
		return res, nil
	}

	// 限制文件上传大小
	maxSize := viper.Init("video").Viper.GetInt("video.maxSizeLimit")
	size := len(req.Data)
	if size > maxSize*1000*1000 {
		logger.Errorln("视频文件过大")
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  fmt.Sprintf("该视频文件大于%dMB，上传受限", maxSize),
		}
		return res, nil
	}

	createTimestamp := time.Now().UnixMilli()
	videoTitle, coverTitle := fmt.Sprintf("%d_%s_%d.mp4", userID, req.Title, createTimestamp), fmt.Sprintf("%d_%s_%d.png", userID, req.Title, createTimestamp)
	// 将视频数据上传至minio
	reader := bytes.NewReader(req.Data)
	contentType := "application/mp4"

	uploadSize, err := minio.UploadFileByIO(minio.VideoBucketName, videoTitle, reader, int64(size), contentType)
	if err != nil {
		logger.Errorf("视频上传至minio失败：%v", err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "视频上传失败",
		}
		return res, nil
	}
	logger.Infof("上传文件大小为:%v", uploadSize)

	// 获取上传文件的路径
	playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, videoTitle)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：视频获取失败",
		}
		return res, nil
	}
	logger.Infof("上传视频路径：%v", playUrl)

	// 截取第一帧并将图像上传至minio
	imgBuffer, err := tool.GetSnapshotImageBuffer(playUrl, 1)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：封面获取失败",
		}
		return res, nil
	}
	var imgByte []byte
	imgBuffer.Write(imgByte)
	contentType = "image/png"

	size = imgBuffer.Len()
	uploadSize, err = minio.UploadFileByIO(minio.CoverBucketName, coverTitle, imgBuffer, int64(size), contentType)
	if err != nil {
		logger.Errorf("封面上传至minio失败：%v", err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：封面上传失败",
		}
		return res, nil
	}
	logger.Infof("上传文件大小为:%v", uploadSize)

	// 获取上传文件的路径
	coverUrl, err := minio.GetFileTemporaryURL(minio.CoverBucketName, coverTitle)
	if err != nil {
		logger.Errorf("封面获取链接失败：%v", err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：封面获取失败",
		}
		return res, nil
	}
	logger.Infof("上传封面路径：%v", coverUrl)

	// 插入数据库
	v := &db.Video{
		Title:    req.Title,
		PlayUrl:  videoTitle,
		CoverUrl: coverTitle,
		AuthorID: uint(userID),
	}
	err = db.CreateVideo(ctx, v)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.PublishActionResponse{
			StatusCode: -1,
			StatusMsg:  "视频发布失败，服务器内部错误",
		}
		return res, nil
	}

	res := &video.PublishActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return res, nil
}

// PublishList implements the VideoServiceImpl interface.
func (s *VideoServiceImpl) PublishList(ctx context.Context, req *video.PublishListRequest) (resp *video.PublishListResponse, err error) {
	logger := zap.InitLogger()
	userID := req.UserId

	results, err := db.GetVideosByUserID(ctx, userID)
	if err != nil {
		logger.Errorln(err.Error())
		res := &video.PublishListResponse{
			StatusCode: -1,
			StatusMsg:  "发布列表获取失败：服务器内部错误",
		}
		return res, nil
	}
	videos := make([]*video.Video, 0)
	for _, r := range results {
		author, err := db.GetUserByID(ctx, int64(r.AuthorID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "发布列表获取失败：服务器内部错误",
			}
			return res, nil
		}
		follow, err := db.GetRelationByUserIDs(ctx, userID, int64(author.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "发布列表获取失败：服务器内部错误",
			}
			return res, nil
		}
		favorite, err := db.GetFavoriteVideoRelationByUserVideoID(ctx, userID, int64(r.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "发布列表获取失败：服务器内部错误",
			}
			return res, nil
		}
		playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, r.PlayUrl)
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：视频获取失败",
			}
			return res, nil
		}
		coverUrl, err := minio.GetFileTemporaryURL(minio.CoverBucketName, r.CoverUrl)
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：封面获取失败",
			}
			return res, nil
		}
		avatarUrl, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, author.Avatar)
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：发布者头像获取失败",
			}
			return res, nil
		}
		backgroundUrl, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, author.BackgroundImage)
		if err != nil {
			logger.Errorln(err.Error())
			res := &video.PublishListResponse{
				StatusCode: -1,
				StatusMsg:  "背景图获取失败",
			}
			return res, nil
		}

		videos = append(videos, &video.Video{
			Id: int64(r.ID),
			Author: &user.User{
				Id:              int64(author.ID),
				Name:            author.UserName,
				FollowerCount:   int64(author.FollowerCount),
				FollowCount:     int64(author.FollowingCount),
				IsFollow:        follow != nil,
				Avatar:          avatarUrl,
				BackgroundImage: backgroundUrl,
				Signature:       author.Signature,
				TotalFavorited:  int64(author.TotalFavorited),
				WorkCount:       int64(author.WorkCount),
				FavoriteCount:   int64(author.FavoriteCount),
			},
			PlayUrl:       playUrl,
			CoverUrl:      coverUrl,
			FavoriteCount: int64(r.FavoriteCount),
			CommentCount:  int64(r.CommentCount),
			IsFavorite:    favorite != nil,
			Title:         r.Title,
		})
	}

	res := &video.PublishListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		VideoList:  videos,
	}
	return res, nil
}
