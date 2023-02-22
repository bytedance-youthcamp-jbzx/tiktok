package response

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/video"
)

type PublishAction struct {
	Base
}

type PublishList struct {
	Base
	VideoList []*video.Video `json:"video_list"`
}

type Feed struct {
	Base
	NextTime  int64          `json:"next_time"`
	VideoList []*video.Video `json:"video_list"`
}
