package service

import (
	"context"
	"fmt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/tool"
	user "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/jwt"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/minio"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
	"math/rand"
	"time"
)

// UserServiceImpl implements the last service interface defined in the IDL.
type UserServiceImpl struct{}

// Register implements the UserServiceImpl interface.
func (s *UserServiceImpl) Register(ctx context.Context, req *user.UserRegisterRequest) (resp *user.UserRegisterResponse, err error) {
	logger := zap.InitLogger()

	// 检查用户名是否冲突
	usr, err := db.GetUserByName(ctx, req.Username)
	if err != nil {
		logger.Errorf("发生错误：%v", err.Error())
		return nil, err
	} else if usr != nil {
		logger.Errorln("该用户名已存在")
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "该用户名已存在，请更换",
		}
		return res, nil
	}

	// 创建user
	rand.Seed(time.Now().UnixMilli())
	if err := db.CreateUsers(ctx, []*db.User{{
		UserName: req.Username,
		Password: tool.Md5Encrypt(req.Password),
		Avatar:   fmt.Sprintf("default%d.png", rand.Intn(10)),
	}}); err != nil {
		logger.Errorf("发生错误：%v", err.Error())
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：注册失败",
		}
		return res, nil
	}

	// 获取用户id
	usr, err = db.GetUserByName(ctx, req.Username)
	if err != nil || usr == nil {
		logger.Errorf("发生错误：%v", err.Error())
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：注册失败",
		}
		return res, nil
	}

	//生成token
	claims := jwt.CustomClaims{Id: int64(usr.ID)}
	claims.ExpiresAt = time.Now().Add(time.Minute * 5).Unix()
	token, err := Jwt.CreateToken(claims)
	if err != nil {
		logger.Errorf("发生错误：%v", err.Error())
		res := &user.UserRegisterResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：token 创建失败",
		}
		return res, nil
	}
	res := &user.UserRegisterResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserId:     int64(usr.ID),
		Token:      token,
	}
	return res, nil
}

// Login implements the UserServiceImpl interface.
func (s *UserServiceImpl) Login(ctx context.Context, req *user.UserLoginRequest) (resp *user.UserLoginResponse, err error) {
	logger := zap.InitLogger()

	// 根据用户名获取密码
	usr, err := db.GetUserByName(ctx, req.Username)
	if err != nil {
		logger.Errorf("发生错误：%v", err.Error())
		return nil, err
	} else if usr == nil {
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "用户名不存在",
		}
		return res, nil
	}

	// 比较数据库中的密码和请求的密码
	if tool.Md5Encrypt(req.Password) != usr.Password {
		logger.Errorln("用户名或密码错误")
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "用户名或密码错误",
		}
		return res, nil
	}

	// 密码认证通过,获取用户id并生成token
	claims := jwt.CustomClaims{
		Id: int64(usr.ID),
	}
	claims.ExpiresAt = time.Now().Add(time.Hour * 24).Unix()
	token, err := Jwt.CreateToken(claims)
	if err != nil {
		logger.Errorf("发生错误：%v", err.Error())
		res := &user.UserLoginResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：token 创建失败",
		}
		return res, nil
	}

	// 返回结果
	res := &user.UserLoginResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		UserId:     int64(usr.ID),
		Token:      token,
	}
	return res, nil
}

// UserInfo implements the UserServiceImpl interface.
func (s *UserServiceImpl) UserInfo(ctx context.Context, req *user.UserInfoRequest) (resp *user.UserInfoResponse, err error) {
	logger := zap.InitLogger()
	userID := req.UserId

	// 从数据库获取user
	usr, err := db.GetUserByID(ctx, userID)
	if err != nil {
		logger.Errorf("发生错误：%v", err.Error())
		res := &user.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：获取背景图失败",
		}
		return res, nil
	}

	avatar, err := minio.GetFileTemporaryURL(minio.AvatarBucketName, usr.Avatar)
	if err != nil {
		logger.Errorf("Minio获取头像失败：%v", err.Error())
		res := &user.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：获取头像失败",
		}
		return res, nil
	}
	backgroundImage, err := minio.GetFileTemporaryURL(minio.BackgroundImageBucketName, usr.BackgroundImage)
	if err != nil {
		logger.Errorf("Minio获取背景图失败：%v", err.Error())
		res := &user.UserInfoResponse{
			StatusCode: -1,
			StatusMsg:  "服务器内部错误：获取背景图失败",
		}
		return res, nil
	}

	//返回结果
	res := &user.UserInfoResponse{
		StatusCode: 0,
		StatusMsg:  "success",
		User: &user.User{
			Id:              int64(usr.ID),
			Name:            usr.UserName,
			FollowCount:     int64(usr.FollowingCount),
			FollowerCount:   int64(usr.FollowerCount),
			IsFollow:        userID == int64(usr.ID),
			Avatar:          avatar,
			BackgroundImage: backgroundImage,
			Signature:       usr.Signature,
			TotalFavorited:  int64(usr.TotalFavorited),
			WorkCount:       int64(usr.WorkCount),
			FavoriteCount:   int64(usr.FavoriteCount),
		},
	}
	return res, nil
}
