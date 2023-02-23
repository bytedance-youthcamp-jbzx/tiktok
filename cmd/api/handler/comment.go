package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bytedance-youthcamp-jbzx/tiktok/cmd/api/rpc"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/response"
	kitex "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/comment"
	"github.com/gin-gonic/gin"
)

func CommentAction(c *gin.Context) {
	token := c.Query("token")
	vid, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.CommentAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "video_id 不合法",
			},
			Comment: nil,
		})
		return
	}
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 64)
	if err != nil || (actionType != 1 && actionType != 2) {
		c.JSON(http.StatusOK, response.CommentAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "action_type 不合法",
			},
			Comment: nil,
		})
		return
	}
	ctx := context.Background()
	req := new(kitex.CommentActionRequest)
	req.Token = token
	req.VideoId = vid
	req.ActionType = int32(actionType)

	if actionType == 1 {
		commentText := c.Query("comment_text")
		if commentText == "" {
			c.JSON(http.StatusOK, response.CommentAction{
				Base: response.Base{
					StatusCode: -1,
					StatusMsg:  "comment_text 不能为空",
				},
				Comment: nil,
			})
			return
		}
		req.CommentText = commentText
	} else if actionType == 2 {
		commentID, err := strconv.ParseInt(c.Query("comment_id"), 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, response.CommentAction{
				Base: response.Base{
					StatusCode: -1,
					StatusMsg:  "comment_id 不合法",
				},
				Comment: nil,
			})
			return
		}
		req.CommentId = commentID
	}
	res, _ := rpc.CommentAction(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.CommentAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
			Comment: nil,
		})
		return
	}
	c.JSON(http.StatusOK, response.CommentAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		Comment: res.Comment,
	})
}

func CommentList(c *gin.Context) {
	token := c.Query("token")
	vid, err := strconv.ParseInt(c.Query("video_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.CommentList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "video_id 不合法",
			},
			CommentList: nil,
		})
		return
	}
	ctx := context.Background()
	req := &kitex.CommentListRequest{
		Token:   token,
		VideoId: vid,
	}
	res, _ := rpc.CommentList(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.CommentList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
			CommentList: nil,
		})
		return
	}
	c.JSON(http.StatusOK, response.CommentList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		CommentList: res.CommentList,
	})
}
