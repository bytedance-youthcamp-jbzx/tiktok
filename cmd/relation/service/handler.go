package service

import (
	"context"
	"encoding/json"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/redis"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/tool"
	relation "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/relation"
	user "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/minio"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
)

// RelationServiceImpl implements the last service interface defined in the IDL.
type RelationServiceImpl struct{}

// RelationAction implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationAction(ctx context.Context, req *relation.RelationActionRequest) (resp *relation.RelationActionResponse, err error) {
	logger := zap.InitLogger()
	// 解析token,获取用户id
	claims, err := Jwt.ParseToken(req.Token)
	if err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "token 解析错误",
		}
		return res, nil
	}
	userID := claims.Id
	toUserID := req.ToUserId
	if req.ActionType != 1 && req.ActionType != 2 {
		logger.Errorln("action_type 格式错误")
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "action_type 格式错误",
		}
		return res, nil
	}
	// 将关注信息存入消息队列，成功存入则表示操作成功，后续处理由redis完成
	relationCache := &redis.RelationCache{
		UserID:     uint(userID),
		ToUserID:   uint(toUserID),
		ActionType: uint(req.ActionType),
		//CreatedAt:  time.Now(),
	}
	jsonRc, _ := json.Marshal(relationCache)
	if err = RelationMq.PublishSimple(ctx, jsonRc); err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationActionResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：操作失败",
		}
		return res, nil
	}
	res := &relation.RelationActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return res, nil
}

// RelationFollowList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowList(ctx context.Context, req *relation.RelationFollowListRequest) (resp *relation.RelationFollowListResponse, err error) {
	logger := zap.InitLogger()
	userID := req.UserId

	// 从数据库获取关注列表
	followings, err := db.GetFollowingListByUserID(ctx, userID)
	if err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationFollowListResponse{
			StatusCode: -1,
			StatusMsg:  "关注列表获取失败",
		}
		return res, nil
	}
	userIDs := make([]int64, 0)
	for _, res := range followings {
		userIDs = append(userIDs, int64(res.ToUserID))
	}
	users, err := db.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationFollowListResponse{
			StatusCode: -1,
			StatusMsg:  "关注列表获取失败",
		}
		return res, nil
	}
	userList := make([]*user.User, 0)
	for _, u := range users {
		avatar, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, u.Avatar)
		if err != nil {
			logger.Errorf("Minio获取头像失败：%v", err.Error())
			res := &relation.RelationFollowListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：头像获取失败",
			}
			return res, nil
		}
		backgroundUrl, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, u.BackgroundImage)
		if err != nil {
			logger.Errorf("Minio获取背景图链接失败：%v", err.Error())
			res := &relation.RelationFollowListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：背景图获取失败",
			}
			return res, nil
		}
		userList = append(userList, &user.User{
			Id:              int64(u.ID),
			Name:            u.UserName,
			FollowCount:     int64(u.FollowingCount),
			FollowerCount:   int64(u.FollowerCount),
			IsFollow:        true,
			Avatar:          avatar,
			BackgroundImage: backgroundUrl,
			Signature:       u.Signature,
			TotalFavorited:  int64(u.TotalFavorited),
			WorkCount:       int64(u.WorkCount),
			FavoriteCount:   int64(u.FavoriteCount),
		})
	}

	// 返回结果
	res := &relation.RelationFollowListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserList:   userList,
	}
	return res, nil
}

// RelationFollowerList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFollowerList(ctx context.Context, req *relation.RelationFollowerListRequest) (resp *relation.RelationFollowerListResponse, err error) {
	logger := zap.InitLogger()
	userID := req.UserId

	// 从数据库获取粉丝列表
	followers, err := db.GetFollowerListByUserID(ctx, userID)
	if err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationFollowerListResponse{
			StatusCode: -1,
			StatusMsg:  "粉丝列表获取失败",
		}
		return res, nil
	}
	userIDs := make([]int64, 0)
	for _, res := range followers {
		userIDs = append(userIDs, int64(res.UserID))
	}
	users, err := db.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationFollowerListResponse{
			StatusCode: -1,
			StatusMsg:  "粉丝列表获取失败",
		}
		return res, nil
	}
	userList := make([]*user.User, 0)
	for _, u := range users {
		// 查询两个用户是否互相关注
		follow, err := db.GetRelationByUserIDs(ctx, userID, int64(u.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &relation.RelationFollowerListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：关系查询失败",
			}
			return res, nil
		}
		avatar, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, u.Avatar)
		if err != nil {
			logger.Errorf("Minio获取头像失败：%v", err.Error())
			res := &relation.RelationFollowerListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：获取头像失败",
			}
			return res, nil
		}
		backgroundUrl, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, u.BackgroundImage)
		if err != nil {
			logger.Errorf("Minio获取背景图链接失败：%v", err.Error())
			res := &relation.RelationFollowerListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：获取背景图失败",
			}
			return res, nil
		}
		userList = append(userList, &user.User{
			Id:              int64(u.ID),
			Name:            u.UserName,
			FollowCount:     int64(u.FollowingCount),
			FollowerCount:   int64(u.FollowerCount),
			IsFollow:        follow != nil,
			Avatar:          avatar,
			BackgroundImage: backgroundUrl,
			Signature:       u.Signature,
			TotalFavorited:  int64(u.TotalFavorited),
			WorkCount:       int64(u.WorkCount),
			FavoriteCount:   int64(u.FavoriteCount),
		})
	}

	// 返回结果
	res := &relation.RelationFollowerListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserList:   userList,
	}
	return res, nil
}

// RelationFriendList implements the RelationServiceImpl interface.
func (s *RelationServiceImpl) RelationFriendList(ctx context.Context, req *relation.RelationFriendListRequest) (resp *relation.RelationFriendListResponse, err error) {
	logger := zap.InitLogger()
	userID := req.UserId

	// 从数据库获取朋友列表
	friends, err := db.GetFriendList(ctx, userID)
	if err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationFriendListResponse{
			StatusCode: -1,
			StatusMsg:  "朋友列表获取失败",
		}
		return res, nil
	}
	userIDs := make([]int64, 0)
	for _, res := range friends {
		userIDs = append(userIDs, int64(res.ToUserID))
	}
	users, err := db.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		logger.Errorln(err.Error())
		res := &relation.RelationFriendListResponse{
			StatusCode: -1,
			StatusMsg:  "朋友列表获取失败",
		}
		return res, nil
	}
	userList := make([]*relation.FriendUser, 0)
	for _, u := range users {
		message, err := db.GetFriendLatestMessage(ctx, userID, int64(u.ID))
		if err != nil {
			res := &relation.RelationFriendListResponse{
				StatusCode: -1,
				StatusMsg:  "获取朋友列表最新消息失败",
			}
			return res, nil
		}
		var msgType int64
		if int64(message.FromUserID) == userID {
			// 当前用户为发送方
			msgType = 1
		} else {
			// 当前用户为接收方
			msgType = 0
		}
		var decContent []byte
		if len(message.Content) != 0 {
			decContent, err = tool.Base64Decode([]byte(message.Content))
			if err != nil {
				logger.Errorf("Base64Decode error: %v\n", err.Error())
				res := &relation.RelationFriendListResponse{
					StatusCode: -1,
					StatusMsg:  "服务器内部错误：获取最新消息失败",
				}
				return res, nil
			}
			decContent, err = tool.RsaDecrypt(decContent, privateKey)
			if err != nil {
				logger.Errorf("rsa decrypt error: %v\n", err.Error())
				res := &relation.RelationFriendListResponse{
					StatusCode: -1,
					StatusMsg:  "服务器内部错误：获取最新消息失败",
				}
				return res, nil
			}
		}
		avatar, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, u.Avatar)
		if err != nil {
			logger.Errorf("Minio获取头像失败：%v", err.Error())
			res := &relation.RelationFriendListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：头像获取失败",
			}
			return res, nil
		}
		backgroundUrl, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, u.BackgroundImage)
		if err != nil {
			logger.Errorf("Minio获取背景图失败：%v", err.Error())
			res := &relation.RelationFriendListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：背景图获取失败",
			}
			return res, nil
		}
		userList = append(userList, &relation.FriendUser{
			Id:              int64(u.ID),
			Name:            u.UserName,
			FollowCount:     int64(u.FollowingCount),
			FollowerCount:   int64(u.FollowerCount),
			IsFollow:        true,
			Message:         string(decContent),
			MsgType:         msgType,
			Avatar:          avatar,
			BackgroundImage: backgroundUrl,
			Signature:       u.Signature,
			TotalFavorited:  int64(u.TotalFavorited),
			WorkCount:       int64(u.WorkCount),
			FavoriteCount:   int64(u.FavoriteCount),
		})
	}

	// 返回结果
	res := &relation.RelationFriendListResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserList:   userList,
	}
	return res, nil
}
