package service

import (
	"context"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/minio"

	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	comment "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/comment"
	user "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"gorm.io/gorm"
)

// CommentServiceImpl implements the last service interface defined in the IDL.
type CommentServiceImpl struct{}

// CommentAction implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentAction(ctx context.Context, req *comment.CommentActionRequest) (resp *comment.CommentActionResponse, err error) {
	logger := zap.InitLogger()
	// 解析token,获取用户id
	claims, err := Jwt.ParseToken(req.Token)
	if err != nil {
		logger.Errorf("token解析错误：%v", err.Error())
		res := &comment.CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "token 解析错误",
		}
		return res, nil
	}
	userID := claims.Id
	actionType := req.ActionType
	v, _ := db.GetVideoById(ctx, req.VideoId)
	if v == nil {
		logger.Errorf("该视频ID不存在：%d", req.VideoId)
		res := &comment.CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "该视频ID不存在",
		}
		return res, nil
	}
	if actionType == 1 {
		cmt := &db.Comment{
			VideoID: uint(req.VideoId),
			UserID:  uint(userID),
			Content: req.CommentText,
		}
		err := db.CreateComment(ctx, cmt)
		if err != nil {
			logger.Errorf("新增评论失败：%v", err.Error())
			res := &comment.CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "评论发布失败：服务器内部错误",
			}
			return res, nil
		}
	} else if actionType == 2 {
		// 判断该评论是否发布自该用户，或该评论在该用户所发布的视频下
		cmt, err := db.GetCommentByCommentID(ctx, req.CommentId)
		if err != nil {
			logger.Errorf("评论删除失败：%v", err.Error())
			res := &comment.CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "评论删除失败：服务器内部错误",
			}
			return res, nil
		}
		if cmt == nil {
			// 评论不存在，无法删除
			logger.Errorf("评论删除失败，该评论ID不存在：%v", req.CommentId)
			res := &comment.CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "评论删除失败：该评论不存在",
			}
			return res, nil
		} else {
			// 查找该视频的作者ID
			v, err := db.GetVideoById(ctx, int64(cmt.VideoID))
			if err != nil {
				logger.Errorf("评论删除失败：%v", err.Error())
				res := &comment.CommentActionResponse{
					StatusCode: -1,
					StatusMsg:  "评论删除失败：服务器内部错误",
				}
				return res, nil
			}
			// 若删除评论的用户不是发布评论的用户或该用户不是视频创作者
			if userID != int64(cmt.UserID) || userID != int64(v.AuthorID) {
				logger.Errorf("评论删除失败，没有权限：%v", cmt.UserID)
				res := &comment.CommentActionResponse{
					StatusCode: -1,
					StatusMsg:  "评论删除失败：没有权限",
				}
				return res, nil
			}
		}
		err = db.DelCommentByID(ctx, req.CommentId, req.VideoId)
		if err != nil {
			logger.Errorf("评论删除失败：%v", err.Error())
			res := &comment.CommentActionResponse{
				StatusCode: -1,
				StatusMsg:  "评论删除失败：服务器内部错误",
			}
			return res, nil
		}
	} else {
		res := &comment.CommentActionResponse{
			StatusCode: -1,
			StatusMsg:  "action_type 非法",
		}
		return res, nil
	}
	res := &comment.CommentActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return res, nil
}

// CommentList implements the CommentServiceImpl interface.
func (s *CommentServiceImpl) CommentList(ctx context.Context, req *comment.CommentListRequest) (resp *comment.CommentListResponse, err error) {
	logger := zap.InitLogger()
	var userID int64 = -1
	// 验证token有效性
	if req.Token != "" {
		claims, err := Jwt.ParseToken(req.Token)
		if err != nil {
			logger.Errorf("token解析错误:%v", err)
			res := &comment.CommentListResponse{
				StatusCode: -1,
				StatusMsg:  "token 解析错误",
			}
			return res, nil
		}
		userID = claims.Id
	}

	// 从数据库获取评论列表
	results, err := db.GetVideoCommentListByVideoID(ctx, req.VideoId)
	if err != nil {
		logger.Errorf("获取评论列表错误：%v", err)
		res := &comment.CommentListResponse{
			StatusCode: -1,
			StatusMsg:  "评论列表获取失败：服务器内部错误",
		}
		return res, nil
	}
	comments := make([]*comment.Comment, 0)
	for _, r := range results {
		u, err := db.GetUserByID(ctx, int64(r.UserID))
		if err != nil && err != gorm.ErrRecordNotFound {
			logger.Errorf("获取用户错误：%v", err.Error())
			res := &comment.CommentListResponse{
				StatusCode: -1,
				StatusMsg:  "评论列表获取失败：服务器内部错误",
			}
			return res, nil
		}
		_, err = db.GetRelationByUserIDs(ctx, userID, int64(u.ID))
		if err != nil {
			logger.Errorln(err.Error())
			res := &comment.CommentListResponse{
				StatusCode: -1,
				StatusMsg:  "评论列表获取失败：服务器内部错误",
			}
			return res, nil
		}
		avatar, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, u.Avatar)
		if err != nil {
			logger.Errorf("Minio获取头像失败：%v", err.Error())
			res := &comment.CommentListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：评论列表用户头像获取失败",
			}
			return res, nil
		}
		backgroundUrl, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, u.Avatar)
		if err != nil {
			logger.Errorf("Minio获取背景图链接失败：%v", err.Error())
			res := &comment.CommentListResponse{
				StatusCode: -1,
				StatusMsg:  "服务器内部错误：评论列表用户背景图获取失败",
			}
			return res, nil
		}
		usr := &user.User{
			Id:              userID,
			Name:            u.UserName,
			FollowCount:     int64(u.FollowingCount),
			FollowerCount:   int64(u.FollowerCount),
			IsFollow:        err != gorm.ErrRecordNotFound,
			Avatar:          avatar,
			BackgroundImage: backgroundUrl,
			Signature:       u.Signature,
			TotalFavorited:  int64(u.TotalFavorited),
			WorkCount:       int64(u.WorkCount),
			FavoriteCount:   int64(u.FavoriteCount),
		}
		comments = append(comments, &comment.Comment{
			Id:         int64(r.ID),
			User:       usr,
			Content:    r.Content,
			CreateDate: r.CreatedAt.Format("2006-01-02"),
			LikeCount:  int64(r.LikeCount),
			TeaseCount: int64(r.TeaseCount),
		})
	}

	res := &comment.CommentListResponse{
		StatusCode:  0,
		StatusMsg:   "success",
		CommentList: comments,
	}
	return res, nil
}
