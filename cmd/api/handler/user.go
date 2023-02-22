// Package handler /*
package handler

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"strconv"

	"github.com/bytedance-youthcamp-jbzx/tiktok/cmd/api/rpc"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/response"
	"github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
	kitex "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
)

// Register 注册
func Register(ctx context.Context, c *app.RequestContext) {
	username := c.Query("username")
	password := c.Query("password")
	//校验参数
	if len(username) == 0 || len(password) == 0 {
		c.JSON(http.StatusBadRequest, response.Register{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "用户名或密码不能为空",
			},
		})
		return
	}
	if len(username) > 32 || len(password) > 32 {
		c.JSON(http.StatusOK, response.Register{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "用户名或密码长度不能大于32个字符",
			},
		})
		return
	}
	//调用kitex/kitex_gen
	req := &kitex.UserRegisterRequest{
		Username: username,
		Password: password,
	}
	res, _ := rpc.Register(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.Register{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.Register{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserID: res.UserId,
		Token:  res.Token,
	})
}

// Login 登录
func Login(ctx context.Context, c *app.RequestContext) {
	username := c.Query("username")
	password := c.Query("password")
	//校验参数
	if len(username) == 0 || len(password) == 0 {
		c.JSON(http.StatusBadRequest, response.Login{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  "用户名或密码不能为空",
			},
		})
		return
	}
	//调用kitex/kitex_gen
	req := &user.UserLoginRequest{
		Username: username,
		Password: password,
	}
	res, _ := rpc.Login(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.Login{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
		})
		return
	}
	c.JSON(http.StatusOK, response.Login{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		UserID: res.UserId,
		Token:  res.Token,
	})
}

// UserInfo 用户信息
func UserInfo(ctx context.Context, c *app.RequestContext) {
	userId := c.Query("user_id")
	token := c.Query("token")
	id, _ := strconv.ParseInt(userId, 10, 64)

	//调用kitex/kitex_genit
	req := &kitex.UserInfoRequest{
		UserId: id,
		Token:  token,
	}
	res, _ := rpc.UserInfo(ctx, req)
	if res.StatusCode == -1 {
		c.JSON(http.StatusOK, response.UserInfo{
			Base: response.Base{
				StatusCode: -1,
				StatusMsg:  res.StatusMsg,
			},
			User: nil,
		})
		return
	}
	c.JSON(http.StatusOK, response.UserInfo{
		Base: response.Base{
			StatusCode: 0,
			StatusMsg:  res.StatusMsg,
		},
		User: res.User,
	})
}
