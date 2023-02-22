package response

import "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/video"

type FavoriteAction struct {
	Base
}

type FavoriteList struct {
	Base
	VideoList []*video.Video `json:"video_list"`
}
