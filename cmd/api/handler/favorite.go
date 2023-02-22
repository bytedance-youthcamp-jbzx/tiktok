package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"strconv"

	"github.com/bytedance-youthcamp-jbzx/tiktok/cmd/api/rpc"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/response"
	kitex "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/favorite"
)

func FavoriteAction(ctx context.Context, c *app.RequestContext) {
	token := c.Query("token")
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 64)
	if err != nil || (actionType != 1 && actionType != 2) {
		c.JSON(http.StatusOK, response.FavoriteAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "action_type 不合法",
			},
		})
		return
	}
	vid, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.FavoriteAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "video_id 不合法",
			},
		})
		return
	}
	req := &kitex.FavoriteActionRequest{
		Token:      token,
		VideoId:    vid,
		ActionType: int32(actionType),
	}
	res, _ := rpc.FavoriteAction(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.FavoriteAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.FavoriteAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
	})
}

func FavoriteList(ctx context.Context, c *app.RequestContext) {
	token := c.Query("token")

	uid, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.FavoriteList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "user_id 不合法",
			},
		})
		return
	}

	req := &kitex.FavoriteListRequest{
		UserId: uid,
		Token:  token,
	}
	res, _ := rpc.FavoriteList(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.FavoriteList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.FavoriteList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		VideoList: res.VideoList,
	})
}
