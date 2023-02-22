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

// CreateVideo
//
//	@Description: 发布一条视频
//	@Date 2023-01-21 16:26:19
//	@param ctx 数据库操作上下文
//	@param video 视频数据
//	@return error
func CreateVideo(ctx context.Context, video *Video) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 在 video 表中创建视频记录
		err := tx.Create(video).Error
		if err != nil {
			return err
		}
		// 2. 同步 user 表中的作品数量
		res := tx.Model(&User{}).Where("id = ?", video.AuthorID).Update("work_count", gorm.Expr("work_count + ?", 1))
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

// GetVideosByUserID
//
//	@Description: 获取用户发布的视频列表
//	@Date 2023-01-21 16:28:44
//	@param ctx 数据库操作上下文
//	@param authorId 作者的用户id
//	@return []*Video 视频列表
//	@return error
func GetVideosByUserID(ctx context.Context, authorId int64) ([]*Video, error) {
	var pubList []*Video
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Model(&Video{}).Where(&Video{AuthorID: uint(authorId)}).Find(&pubList).Error
	if err != nil {
		return nil, err
	}
	return pubList, nil
}

// DelVideoByID
//
//	@Description: 根据视频id和作者id删除视频
//	@Date 2023-02-22 23:34:45
//	@param ctx 数据库操作上下文
//	@param videoID 视频id
//	@param authorID 作者id
//	@return error
func DelVideoByID(ctx context.Context, videoID int64, authorID int64) error {
	err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. 根据主键 video_id 删除 video
		err := tx.Unscoped().Delete(&Video{}, videoID).Error
		if err != nil {
			return err
		}
		// 2. 同步 user 表中的作品数量
		res := tx.Model(&User{}).Where("id = ?", authorID).Update("work_count", gorm.Expr("work_count - ?", 1))
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
