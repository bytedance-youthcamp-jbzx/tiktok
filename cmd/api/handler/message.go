package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"strconv"

	"github.com/bytedance-youthcamp-jbzx/tiktok/cmd/api/rpc"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/response"
	kitex "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/message"
)

func MessageChat(ctx context.Context, c *app.RequestContext) {
	token := c.Query("token")
	toUserID, err := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.MessageChat{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "to_user_id 不合法",
			},
			MessageList: nil,
		})
		return
	}

	// 调用kitex/kitex_gen
	req := &kitex.MessageChatRequest{
		Token:    token,
		ToUserId: toUserID,
	}
	res, _ := rpc.MessageChat(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.MessageChat{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
			MessageList: nil,
		})
		return
	}
	c.JSON(http.StatusOK, response.MessageChat{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		MessageList: res.MessageList,
	})
}

func MessageAction(ctx context.Context, c *app.RequestContext) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusOK, response.RelationAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "Token has expired.",
			},
		})
		return
	}

	toUserID, err := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.MessageAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "to_user_id 不合法",
			},
		})
		return
	}
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 64)
	if err != nil || actionType != 1 {
		c.JSON(http.StatusOK, response.MessageAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "action_type 不合法",
			},
		})
		return
	}

	// 调用kitex/kitex_gen
	req := &kitex.MessageActionRequest{
		Token:      token,
		ToUserId:   toUserID,
		ActionType: int32(actionType),
		Content:    c.Query("content"),
	}
	res, _ := rpc.MessageAction(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.MessageAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.MessageAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
	})
}
