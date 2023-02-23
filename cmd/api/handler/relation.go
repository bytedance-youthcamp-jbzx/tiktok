package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/bytedance-youthcamp-jbzx/tiktok/cmd/api/rpc"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/response"
	kitex "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/relation"
	"github.com/gin-gonic/gin"
)

func FriendList(c *gin.Context) {
	token := c.Query("token")
	uid, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.FriendList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "user_id 不合法",
			},
		})
		return
	}
	// 调用kitex/kitex_gen
	ctx := context.Background()
	req := &kitex.RelationFriendListRequest{
		UserId: uid,
		Token:  token,
	}
	res, _ := rpc.RelationFriendList(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.FriendList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.FriendList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  "success",
		},
		UserList: res.UserList,
	})
}

func FollowerList(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.FollowerList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "user_id 不合法",
			},
		})
		return
	}

	ctx := context.Background()
	token := c.Query("token")
	req := &kitex.RelationFollowerListRequest{
		UserId: uid,
		Token:  token,
	}
	res, _ := rpc.RelationFollowerList(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.FollowerList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.FollowerList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserList: res.UserList,
	})
}

func FollowList(c *gin.Context) {
	uid, err := strconv.ParseInt(c.Query("user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.FollowList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "user_id 不合法",
			},
		})
		return
	}
	ctx := context.Background()
	token := c.Query("token")
	req := &kitex.RelationFollowListRequest{
		UserId: uid,
		Token:  token,
	}
	res, _ := rpc.RelationFollowList(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.FollowList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.FollowList{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserList: res.UserList,
	})
}

func RelationAction(c *gin.Context) {
	tid, err := strconv.ParseInt(c.Query("to_user_id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, response.RelationAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "to_user_id 不合法",
			},
		})
		return
	}
	actionType, err := strconv.ParseInt(c.Query("action_type"), 10, 64)
	if err != nil || (actionType != 1 && actionType != 2) {
		c.JSON(http.StatusOK, response.RelationAction{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "action_type 不合法",
			},
		})
		return
	}
	ctx := context.Background()
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
	req := &kitex.RelationActionRequest{
		Token:      token,
		ToUserId:   tid,
		ActionType: int32(actionType),
	}
	res, _ := rpc.RelationAction(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.FollowList{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.RelationAction{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
	})
}
