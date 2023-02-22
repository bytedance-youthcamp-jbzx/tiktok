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
	"time"

	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/errno"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// Comment
//
//	@Description: 用户评论数据模型
type Comment struct {
	ID         uint      `gorm:"primarykey"`
	CreatedAt  time.Time `gorm:"index;not null" json:"create_date"`
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	Video      Video          `gorm:"foreignkey:VideoID" json:"video,omitempty"`
	VideoID    uint           `gorm:"index:idx_videoid;not null" json:"video_id"`
	User       User           `gorm:"foreignkey:UserID" json:"user,omitempty"`
	UserID     uint           `gorm:"index:idx_userid;not null" json:"user_id"`
	Content    string         `gorm:"type:varchar(255);not null" json:"content"`
	LikeCount  uint           `gorm:"column:like_count;default:0;not null" json:"like_count,omitempty"`
	TeaseCount uint           `gorm:"column:tease_count;default:0;not null" json:"tease_count,omitempty"`
}

func (Comment) TableName() string {
	return "comments"
}

// CreateComment
//
//	@Description: 新增一条评论数据，并对所属视频的评论数+1
//	@Date 2023-01-21 14:42:49
//	@param ctx 数据库操作上下文
//	@param comment 评论数据
//	@return error
func CreateComment(ctx context.Context, comment *Comment) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		// 1. 新增评论数据
		err := tx.Create(comment).Error
		if err != nil {
			return err
		}

		// 2.对 Video 表中的评论数+1
		res := tx.Model(&Video{}).Where("id = ?", comment.VideoID).Update("comment_count", gorm.Expr("comment_count + ?", 1))
		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected != 1 {
			// 影响的数据条数不是1
			return errno.ErrDatabase
		}

		return nil
	})
	return err
}

// DelCommentByID
//
//	@Description: 删除一条评论数据，并对所属视频的评论数-1
//	@Date 2023-01-21 14:49:43
//	@param ctx 数据库操作上下文
//	@param commentID 需要删除的评论的id
//	@param vid 评论所属视频的id
//	@return error
func DelCommentByID(ctx context.Context, commentID int64, vid int64) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 在事务中执行一些 db 操作（从这里开始，您应该使用 'tx' 而不是 'db'）
		comment := new(Comment)
		if err := tx.First(&comment, commentID).Error; err != nil {
			return err
		} else if err == gorm.ErrRecordNotFound {
			return nil
		}

		// 1. 删除评论数据
		// 这里使用的实际上是软删除
		err := tx.Where("id = ?", commentID).Delete(&Comment{}).Error
		if err != nil {
			return err
		}

		// 2.改变 video 表中的 comment count
		res := tx.Model(&Video{}).Where("id = ?", vid).Update("comment_count", gorm.Expr("comment_count - ?", 1))
		if res.Error != nil {
			return res.Error
		}

		if res.RowsAffected != 1 {
			// 影响的数据条数不是1
			return errno.ErrDatabase
		}

		return nil
	})
	return err
}

// GetVideoCommentListByVideoID
//
//	@Description: 根据视频id获取指定视频的全部评论内容
//	@Date 2023-01-21 15:13:33
//	@param ctx 数据库操作上下文
//	@param videoID 视频id
//	@return []*Comment 评论内容
//	@return error
func GetVideoCommentListByVideoID(ctx context.Context, videoID int64) ([]*Comment, error) {
	var comments []*Comment
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Model(&Comment{}).Where(&Comment{VideoID: uint(videoID)}).Order("created_at DESC").Find(&comments).Error
	if err != nil {
		return nil, err
	}
	return comments, nil
}
