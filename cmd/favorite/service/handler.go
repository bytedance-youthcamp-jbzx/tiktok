package service

import (
	"context"
	"encoding/json"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/redis"
	favorite "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/favorite"
	user "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
	video "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/video"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/minio"
)

// FavoriteServiceImpl implements the last service interface defined in the IDL.
type FavoriteServiceImpl struct{}

// FavoriteAction implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteAction(ctx context.Context, req *favorite.FavoriteActionRequest) (resp *favorite.FavoriteActionResponse, err error) {
	// 解析token,获取用户id
	claims, err := Jwt.ParseToken(req.Token)
	if err != nil {
		logger.Errorf("token解析错误：%v", err.Error())
		res := &favorite.FavoriteActionResponse{
			StatusCode: -1,
			StatusMsg:  "token 解析错误",
		}
		return res, nil
	}
	userID := claims.Id

	//将点赞信息存入消息队列,成功存入则表示点赞成功,后续处理由redis完成
	fc := &redis.FavoriteCache{
		VideoID:    uint(req.VideoId),
		UserID:     uint(userID),
		ActionType: uint(req.ActionType),
		//CreatedAt:  time.Now(),
	}
	jsonFC, _ := json.Marshal(fc)
	if err = FavoriteMq.PublishSimple(ctx, jsonFC); err != nil {
		logger.Errorf("消息队列发布错误:%v", err.Error())
		res := &favorite.FavoriteActionResponse{
			StatusCode: -1,
			StatusMsg:  "操作失败：服务器内部错误",
		}
		return res, nil
	}
	res := &favorite.FavoriteActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return res, nil
}

// FavoriteList implements the FavoriteServiceImpl interface.
func (s *FavoriteServiceImpl) FavoriteList(ctx context.Context, req *favorite.FavoriteListRequest) (resp *favorite.FavoriteListResponse, err error) {
	userID := req.UserId

	// 从数据库获取喜欢列表
	results, err := db.GetFavoriteListByUserID(ctx, userID)
	if err != nil {
		logger.Errorf("获取喜欢列表错误：%v", err.Error())
		res := &favorite.FavoriteListResponse{
			StatusCode: -1,
			StatusMsg:  "获取喜欢列表失败：服务器内部错误",
		}
		return res, nil
	}
	favorites := make([]*video.Video, 0)
	for _, r := range results {
		v, err := db.GetVideoById(ctx, int64(r.VideoID))
		if err != nil {
			logger.Errorf("获取视频错误：%v", err.Error())
			res := &favorite.FavoriteListResponse{
				StatusCode: -1,
				StatusMsg:  "获取喜欢列表失败：服务器内部错误",
			}
			return res, nil
		}

		u, err := db.GetUserByID(ctx, int64(v.AuthorID))
		if err != nil {
			logger.Errorf("获取用户错误：%v", err.Error())
			res := &favorite.FavoriteListResponse{
				StatusCode: -1,
				StatusMsg:  "获取喜欢列表失败：服务器内部错误",
			}
			return res, nil
		}

		relation, err := db.GetRelationByUserIDs(ctx, userID, int64(u.ID))
		if err != nil {
			logger.Errorf("发生错误：%v", err.Error())
			res := &favorite.FavoriteListResponse{
				StatusCode: -1,
				StatusMsg:  "获取喜欢列表失败：服务器内部错误",
			}
			return res, nil
		}
		playUrl, err := minio.GetFileTemporaryURL(minio.VideoBucketName, v.PlayUrl)
		if err != nil {
			logger.Errorf("发生错误：%v", err.Error())
			res := &favorite.FavoriteListResponse{
				StatusCode: -1,
				StatusMsg:  "获取喜欢列表失败：服务器内部错误",
			}
			return res, nil
		}
		coverUrl, err := minio.GetFileTemporaryURL(minio.CoverBucketName, v.CoverUrl)
		if err != nil {
			logger.Errorf("发生错误：%v", err.Error())
			res := &favorite.FavoriteListResponse{
				StatusCode: -1,
				StatusMsg:  "获取喜欢列表失败：服务器内部错误",
			}
			return res, nil
		}
		avatar, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, u.Avatar)
		if err != nil {
			logger.Errorf("Minio获取头像失败：%v", err.Error())
			res := &favorite.FavoriteListResponse{
				StatusCode: -1,
				StatusMsg:  "获取喜欢列表失败：服务器内部错误",
			}
			return res, nil
		}
		backgroundUrl, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, u.BackgroundImage)
		if err != nil {
			logger.Errorf("Minio获取链接失败：%v", err.Error())
			res := &favorite.FavoriteListResponse{
				StatusCode: -1,
				StatusMsg:  "获取喜欢列表失败：服务器内部错误",
			}
			return res, nil
		}
		favorites = append(favorites, &video.Video{
			Id: int64(r.VideoID),
			Author: &user.User{
				Id:              int64(u.ID),
				Name:            u.UserName,
				FollowCount:     int64(u.FollowingCount),
				FollowerCount:   int64(u.FollowerCount),
				IsFollow:        relation != nil,
				Avatar:          avatar,
				BackgroundImage: backgroundUrl,
				Signature:       u.Signature,
				TotalFavorited:  int64(u.TotalFavorited),
				WorkCount:       int64(u.WorkCount),
				FavoriteCount:   int64(u.FavoriteCount),
			},
			PlayUrl:       playUrl,
			CoverUrl:      coverUrl,
			FavoriteCount: int64(v.FavoriteCount),
			CommentCount:  int64(v.CommentCount),
			IsFavorite:    true,
			Title:         v.Title,
		})
	}

	res := &favorite.FavoriteListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		VideoList:  favorites,
	}
	return res, nil
}
