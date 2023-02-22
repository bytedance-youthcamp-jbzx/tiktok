package db

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

// Message
//
//	@Description: 聊天消息数据模型
type Message struct {
	ID         uint      `gorm:"primarykey"`
	CreatedAt  time.Time `gorm:"index;not null" json:"create_time"`
	UpdatedAt  time.Time
	DeletedAt  gorm.DeletedAt `gorm:"index"`
	FromUser   User           `gorm:"foreignkey:FromUserID;" json:"from_user,omitempty"`
	FromUserID uint           `gorm:"index:idx_userid_from;not null" json:"from_user_id"`
	ToUser     User           `gorm:"foreignkey:ToUserID;" json:"to_user,omitempty"`
	ToUserID   uint           `gorm:"index:idx_userid_from;index:idx_userid_to;not null" json:"to_user_id"`
	Content    string         `gorm:"type:varchar(255);not null" json:"content"`
}

func (Message) TableName() string {
	return "messages"
}

// GetMessagesByUserIDs
//
//		@Description: 根据两个用户的用户id获取聊天信息记录
//		@Date 2023-01-25 11:37:08
//		@param ctx 数据库操作上下文
//		@param userID 主用户id
//		@param toUserID 对象用户id
//	 @param lastTimestamp 要查询消息时间的下限
//		@return []*Message 聊天信息数据列表
//		@return error
func GetMessagesByUserIDs(ctx context.Context, userID int64, toUserID int64, lastTimestamp int64) ([]*Message, error) {
	res := make([]*Message, 0)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("((from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)) AND created_at > ?",
		userID, toUserID, toUserID, userID, time.UnixMilli(lastTimestamp).Format("2006-01-02 15:04:05.000"),
	).Order("created_at ASC").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// GetMessagesByUserToUser
//
//	@Description: 根据两个用户的用户id获取单向数据
//	@Date 2023-01-25 11:37:08
//	@param ctx 数据库操作上下文
//	@param userID 主用户id
//	@param toUserID 对象用户id
//  @param lastTimestamp 要查询消息时间的下限
//	@return []*Message 聊天信息数据列表
//	@return error

func GetMessagesByUserToUser(ctx context.Context, userID int64, toUserID int64, lastTimestamp int64) ([]*Message, error) {
	res := make([]*Message, 0)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Where("from_user_id = ? AND to_user_id = ? AND created_at > ?",
		userID, toUserID, time.UnixMilli(lastTimestamp).Format("2006-01-02 15:04:05.000"),
	).Order("created_at ASC").Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// CreateMessagesByList
//
//	@Description: 新增多条聊天信息
//	@Date 2023-01-21 17:13:26
//	@param ctx 数据库操作上下文
//	@param users 用户数据列表
//	@return error
func CreateMessagesByList(ctx context.Context, messages []*Message) error {
	err := GetDB().Clauses(dbresolver.Write).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(messages).Error; err != nil {
			return err
		}
		return nil
	})
	return err
}

// GetMessageIDsByUserIDs
//
//	@Description: 查询消息ID
//	@Date 2023-01-21 17:13:26
//	@param ctx 数据库操作上下文
//	@param userID 发送用户
//	@param toUserID 接收用户
//	@return error
func GetMessageIDsByUserIDs(ctx context.Context, userID int64, toUserID int64) ([]*Message, error) {
	res := make([]*Message, 0)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Select("id").Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)", userID, toUserID, toUserID, userID).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// GetMessageByID
//
//	@Description: 根据消息ID查询消息
//	@Date 2023-01-21 17:13:26
//	@param ctx 数据库操作上下文
//	@param messageID 消息ID列表
//	@return error
func GetMessageByID(ctx context.Context, messageID int64) (*Message, error) {
	res := new(Message)
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Select("id, from_user_id, to_user_id, content, created_at").Where("id = ?", messageID).First(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

func GetFriendLatestMessage(ctx context.Context, userID int64, toUserID int64) (*Message, error) {
	var res *Message
	if err := GetDB().Clauses(dbresolver.Read).WithContext(ctx).Select("id, from_user_id, to_user_id, content, created_at").Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)", userID, toUserID, toUserID, userID).Order("created_at DESC").Limit(1).Find(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}
