package response

import (
	"github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/relation"
	"github.com/bytedance-youthcamp-jbzx/tiktok/kitex/kitex_gen/user"
)

type RelationAction struct {
	Base
}

type FollowerList struct {
	Base
	UserList []*user.User `json:"user_list"`
}

type FollowList struct {
	Base
	UserList []*user.User `json:"user_list"`
}

type FriendList struct {
	Base
	UserList []*relation.FriendUser `json:"user_list"`
}
