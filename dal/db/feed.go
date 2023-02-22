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

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// Video
//
//	@Description: 视频数据模型
type Video struct {
	ID            uint      `gorm:"primarykey"`
	CreatedAt     time.Time `gorm:"not null;index:idx_create" json:"created_at,omitempty"`
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
	Author        User           `gorm:"foreignkey:AuthorID" json:"author,omitempty"`
	AuthorID      uint           `gorm:"index:idx_authorid;not null" json:"author_id,omitempty"`
	PlayUrl       string         `gorm:"type:varchar(255);not null" json:"play_url,omitempty"`
	CoverUrl      string         `gorm:"type:varchar(255)" json:"cover_url,omitempty"`
	FavoriteCount uint           `gorm:"default:0;not null" json:"favorite_count,omitempty"`
	CommentCount  uint           `gorm:"default:0;not null" json:"comment_count,omitempty"`
	Title         string         `gorm:"type:varchar(50);not null" json:"title,omitempty"`
}

func (Video) TableName() string {
	return "videos"
}

// MGetVideos
//
//	@Description: 获取最近发布的视频
//	@Date 2023-01-21 16:39:00
//	@param ctx
//	@param limit 获取的视频条数
//	@param latestTime 最早的时间限制
//	@return []*Video 视频列表
//	@return error
func MGetVideos(ctx context.Context, limit int, latestTime *int64) ([]*Video, error) {
	videos := make([]*Video, 0)

	if latestTime == nil || *latestTime == 0 {
		curTime := time.Now().UnixMilli()
		latestTime = &curTime
	}
	conn := GetDB().Clauses(dbresolver.Read).WithContext(ctx)
	if err := conn.Limit(limit).Order("created_at desc").Find(&videos, "created_at < ?", time.UnixMilli(*latestTime)).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

// GetVideoById
//
//	@Description: 根据视频id获取视频
//	@Date 2023-01-24 15:58:52
//	@param ctx 数据库操作上下文
//	@param videoID 视频id
//	@return *Video 视频数据
//	@return error
func GetVideoById(ctx context.Context, videoID int64) (*Video, error) {
	res := new(Video)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).First(&res, videoID).Error; err == nil {
		return res, nil
	} else if err == gorm.ErrRecordNotFound {
		return nil, nil
	} else {
		return nil, err
	}
}

// GetVideoListByIDs
//
//	@Description: 根据视频id列表获取视频列表
//	@Date 2023-01-24 16:00:12
//	@param ctx 数据库操作上下文
//	@param videoIDs 视频id列表
//	@return []*Video 视频数据列表
//	@return error
func GetVideoListByIDs(ctx context.Context, videoIDs []int64) ([]*Video, error) {
	res := make([]*Video, 0)
	if len(videoIDs) == 0 {
		return res, nil
	}

	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("video_id in ?", videoIDs).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
