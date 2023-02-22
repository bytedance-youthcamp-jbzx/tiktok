//
// Package db
// @Description: 数据库数据库操作业务逻辑
// @Author hehehhh
// @Date 2023-01-21 14:33:47
// @Update
//

package db

import (
	"context"

	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/errno"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// FavoriteVideoRelation
//
//	@Description: 用户与视频的点赞关系数据模型
type FavoriteVideoRelation struct {
	Video   Video `gorm:"foreignkey:VideoID;" json:"video,omitempty"`
	VideoID uint  `gorm:"index:idx_videoid;not null" json:"video_id"`
	User    User  `gorm:"foreignkey:UserID;" json:"user,omitempty"`
	UserID  uint  `gorm:"index:idx_userid;not null" json:"user_id"`
}

// FavoriteCommentRelation
//
//	@Description: 用户与评论的点赞关系数据模型
type FavoriteCommentRelation struct {
	Comment   Comment `gorm:"foreignkey:CommentID;" json:"comment,omitempty"`
	CommentID uint    `gorm:"column:comment_id;index:idx_commentid;not null" json:"video_id"`
	User      User    `gorm:"foreignkey:UserID;" json:"user,omitempty"`
	UserID    uint    `gorm:"column:user_id;index:idx_userid;not null" json:"user_id"`
}

func (FavoriteVideoRelation) TableName() string {
	return "user_favorite_videos"
}

func (FavoriteCommentRelation) TableName() string {
	return "user_favorite_comments"
}

// CreateVideoFavorite
//
//	@Description: 创建一条用户点赞数据
//	@Date 2023-01-21 17:19:20
//	@param ctx 数据库操作上下文
//	@param userID 用户id
//	@param videoID 视频id
//	@param authorID 视频作者id
//	@return error
func CreateVideoFavorite(ctx context.Context, userID int64, videoID int64, authorID int64) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		//1. 新增点赞数据
		err := tx.Create(&FavoriteVideoRelation{UserID: uint(userID), VideoID: uint(videoID)}).Error
		if err != nil {
			return err
		}

		//2.改变 video 表中的 favorite count
		res := tx.Model(&Video{}).Where("id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected != 1 {
			// 影响的数据条数不是1
			return errno.ErrDatabase
		}

		//3.改变 user 表中的 favorite count
		res = tx.Model(&User{}).Where("id = ?", userID).Update("favorite_count", gorm.Expr("favorite_count + ?", 1))
		if res.Error != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return errno.ErrDatabase
		}

		//4.改变 user 表中的 total_favorited
		res = tx.Model(&User{}).Where("id = ?", authorID).Update("total_favorited", gorm.Expr("total_favorited + ?", 1))
		if res.Error != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return errno.ErrDatabase
		}

		return nil
	})
	return err
}

// GetFavoriteVideoRelationByUserVideoID
//
//	@Description: 获取用户与视频之间的点赞关系
//	@Date 2023-01-21 16:49:38
//	@param ctx 数据库操作上下文
//	@param userID 用户id
//	@param videoID 视频id
//	@return *Video 视频数据
//	@return error
func GetFavoriteVideoRelationByUserVideoID(ctx context.Context, userID int64, videoID int64) (*FavoriteVideoRelation, error) {
	FavoriteVideoRelation := new(FavoriteVideoRelation)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).First(&FavoriteVideoRelation, "user_id = ? and video_id = ?", userID, videoID).Error; err == nil {
		return FavoriteVideoRelation, nil
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	} else {
		return nil, err
	}
}

// DelFavoriteByUserVideoID
//
//	@Description: 删除用户对视频的点赞信息，并对所属视频的点赞数-1
//	@Date 2023-01-21 16:05:11
//	@param ctx 数据库操作上下文
//	@param userID 用户id
//	@param videoID 视频id
//	@param authorID 视频作者id
//	@return error
func DelFavoriteByUserVideoID(ctx context.Context, userID int64, videoID int64, authorID int64) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		FavoriteVideoRelation := new(FavoriteVideoRelation)
		if err := tx.Where("user_id = ? and video_id = ?", userID, videoID).First(&FavoriteVideoRelation).Error; err != nil {
			return err
		} else if err == gorm.ErrRecordNotFound {
			return nil
		}

		//1. 删除点赞数据
		// 因为FavoriteVideoRelation中包含了gorm.Model所以拥有软删除能力
		// 而tx.Unscoped().Delete()将永久删除记录
		err := tx.Unscoped().Where("user_id = ? and video_id = ?", userID, videoID).Delete(&FavoriteVideoRelation).Error
		if err != nil {
			return err
		}

		//2.改变 video 表中的 favorite count
		res := tx.Model(&Video{}).Where("id = ?", videoID).Update("favorite_count", gorm.Expr("favorite_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected != 1 {
			// 影响数据条数不是1
			return errno.ErrDatabase
		}

		//3.改变 user 表中的 favorite count
		res = tx.Model(&User{}).Where("id = ?", userID).Update("favorite_count", gorm.Expr("favorite_count - ?", 1))
		if res.Error != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return errno.ErrDatabase
		}

		//4.改变 user 表中的 total_favorited
		res = tx.Model(&User{}).Where("id = ?", authorID).Update("total_favorited", gorm.Expr("total_favorited - ?", 1))
		if res.Error != nil {
			return err
		}
		if res.RowsAffected != 1 {
			return errno.ErrDatabase
		}

		return nil
	})
	return err
}

// GetFavoriteListByUserID
//
//	@Description: 根据用户id获取用户的点赞关系列表
//	@Date 2023-01-21 17:08:52
//	@param ctx 数据库操作上下文
//	@param userID 用户id
//	@return []*FavoriteVideoRelation 点赞关系列表
//	@return error
func GetFavoriteListByUserID(ctx context.Context, userID int64) ([]*FavoriteVideoRelation, error) {
	var FavoriteVideoRelationList []*FavoriteVideoRelation
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("user_id = ?", userID).Find(&FavoriteVideoRelationList).Error
	if err != nil {
		return nil, err
	}
	return FavoriteVideoRelationList, nil
}

// GetAllFavoriteList
//
//	@Description: 获取全部的点赞关系列表
//	@Date 2023-02-17 16:15:00
//	@param ctx 数据库操作上下文
//	@return []*FavoriteVideoRelation 所有的点赞关系列表
//	@return error
func GetAllFavoriteList(ctx context.Context) ([]*FavoriteVideoRelation, error) {
	var FavoriteVideoRelationList []*FavoriteVideoRelation
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Find(&FavoriteVideoRelationList).Error; err != nil {
		return nil, err
	}
	return FavoriteVideoRelationList, nil
}

// GetFavoriteCommentRelationByUserCommentID
//
//	@Description: 获取评论的点赞关系
//	@Date 2023-02-17 21:28:00
//	@param ctx 数据库操作上下文
//	@param userID 用户ID
//	@param commentID 评论ID
//	@return *FavoriteCommentRelation 评论点赞关系
//	@return error
func GetFavoriteCommentRelationByUserCommentID(ctx context.Context, userID int64, commentID int64) (*FavoriteCommentRelation, error) {
	FavoriteCommentRelation := new(FavoriteCommentRelation)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).First(&FavoriteCommentRelation, "user_id = ? and comment_id = ?", userID, commentID).Error; err == nil {
		return FavoriteCommentRelation, nil
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	} else {
		return nil, err
	}
}
