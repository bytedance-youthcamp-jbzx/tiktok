package service

import (
	"context"

	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/db"
	"github.com/bytedance-youthcamp-jbzx/tiktok/dal/redis"
	"github.com/bytedance-youthcamp-jbzx/tiktok/internal/tool"
	message "github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/message"
	"github.com/bytedance-youthcamp-jbzx/tiktok/pkg/zap"
)

// MessageServiceImpl implements the last service interface defined in the IDL.
type MessageServiceImpl struct{}

// MessageChat implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MessageChat(ctx context.Context, req *message.MessageChatRequest) (resp *message.MessageChatResponse, err error) {
	logger := zap.InitLogger()
	// 解析token,获取用户id
	claims, err := Jwt.ParseToken(req.Token)
	if err != nil {
		logger.Errorln(err.Error())
		res := &message.MessageChatResponse{
			StatusCode: -1,
			StatusMsg:  "token 解析错误",
		}
		return res, nil
	}
	userID := claims.Id

	// 从redis中获取message时间戳
	lastTimestamp, err := redis.GetMessageTimestamp(ctx, req.Token, req.ToUserId)
	if err != nil {
		logger.Errorln(err.Error())
		res := &message.MessageChatResponse{
			StatusCode: -1,
			StatusMsg:  "聊天记录获取失败：服务器内部错误",
		}
		return res, nil
	}

	var results []*db.Message
	if lastTimestamp == -1 {
		results, err = db.GetMessagesByUserIDs(ctx, userID, req.ToUserId, int64(lastTimestamp))
		lastTimestamp = 0
	} else {
		results, err = db.GetMessagesByUserToUser(ctx, req.ToUserId, userID, int64(lastTimestamp))
	}

	if err != nil {
		logger.Errorln(err.Error())
		res := &message.MessageChatResponse{
			StatusCode: -1,
			StatusMsg:  "聊天记录获取失败：服务器内部错误",
		}
		return res, nil
	}
	messages := make([]*message.Message, 0)
	for _, r := range results {
		decContent, err := tool.Base64Decode([]byte(r.Content))
		if err != nil {
			logger.Errorf("Base64Decode error: %v\n", err.Error())
			res := &message.MessageChatResponse{
				StatusCode: -1,
				StatusMsg:  "聊天记录获取失败：服务器内部错误",
			}
			return res, nil
		}
		decContent, err = tool.RsaDecrypt(decContent, privateKey)
		if err != nil {
			logger.Errorf("rsa decrypt error: %v\n", err.Error())
			res := &message.MessageChatResponse{
				StatusCode: -1,
				StatusMsg:  "聊天记录获取失败：服务器内部错误",
			}
			return res, nil
		}
		messages = append(messages, &message.Message{
			Id:         int64(r.ID),
			FromUserId: int64(r.FromUserID),
			ToUserId:   int64(r.ToUserID),
			Content:    string(decContent),
			CreateTime: r.CreatedAt.UnixMilli(),
		})
	}

	res := &message.MessageChatResponse{
		StatusCode:  0,
		StatusMsg:   "success",
		MessageList: messages,
	}

	// 更新时间redis里的时间戳
	if len(messages) > 0 {
		message := messages[len(messages)-1]
		lastTimestamp = int(message.CreateTime)
	}

	if err = redis.SetMessageTimestamp(ctx, req.Token, req.ToUserId, lastTimestamp); err != nil {
		logger.Errorln(err.Error())
		res := &message.MessageChatResponse{
			StatusCode: -1,
			StatusMsg:  "聊天记录获取失败：服务器内部错误",
		}
		return res, nil
	}

	return res, nil
}

// MessageAction implements the MessageServiceImpl interface.
func (s *MessageServiceImpl) MessageAction(ctx context.Context, req *message.MessageActionRequest) (resp *message.MessageActionResponse, err error) {
	logger := zap.InitLogger()
	// 解析token,获取用户id
	claims, err := Jwt.ParseToken(req.Token)
	if err != nil {
		logger.Errorln(err.Error())
		res := &message.MessageActionResponse{
			StatusCode: -1,
			StatusMsg:  "token 解析错误",
		}
		return res, nil
	}
	userID := claims.Id

	toUserID, actionType := req.ToUserId, req.ActionType

	if userID == toUserID {
		logger.Errorln("不能给自己发送消息")
		res := &message.MessageActionResponse{
			StatusCode: -1,
			StatusMsg:  "消息发送失败：不能给自己发送消息",
		}
		return res, nil
	}

	relation, err := db.GetRelationByUserIDs(ctx, userID, toUserID)
	if relation == nil {
		logger.Errorf("消息发送失败：非朋友关系，无法发送")
		res := &message.MessageActionResponse{
			StatusCode: -1,
			StatusMsg:  "消息发送失败：非朋友关系，无法发送",
		}
		return res, nil
	}

	rsaContent, err := tool.RsaEncrypt([]byte(req.Content), publicKey)
	if err != nil {
		logger.Errorf("rsa encrypt error: %v\n", err.Error())
		res := &message.MessageActionResponse{
			StatusCode: -1,
			StatusMsg:  "消息发送失败：服务器内部错误",
		}
		return res, nil
	}

	messages := make([]*db.Message, 0)
	messages = append(messages, &db.Message{
		FromUserID: uint(userID),
		ToUserID:   uint(toUserID),
		Content:    string(tool.Base64Encode(rsaContent)),
	})
	if actionType == 1 {
		err := db.CreateMessagesByList(ctx, messages)
		if err != nil {
			logger.Errorln(err.Error())
			res := &message.MessageActionResponse{
				StatusCode: -1,
				StatusMsg:  "消息发送失败：服务器内部错误",
			}
			return res, nil
		}
	} else {
		logger.Errorf("action_type 非法：%v", actionType)
		res := &message.MessageActionResponse{
			StatusCode: -1,
			StatusMsg:  "消息发送失败：非法的 action_type",
		}
		return res, nil
	}
	res := &message.MessageActionResponse{
		StatusCode: 0,
		StatusMsg:  "success",
	}
	return res, nil
}
