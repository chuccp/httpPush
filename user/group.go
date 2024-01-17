package user

import "time"

type Group struct {
	createTime   *time.Time
	lastLiveTime *time.Time
}

func NewGroup(user IUser) *Group {
	return &Group{createTime: user.LastLiveTime(), lastLiveTime: user.LastLiveTime()}
}

type GroupUser struct {
	usernames map[string]any
}

func NewGroupUser() *GroupUser {
	return &GroupUser{usernames: make(map[string]any)}
}

var _myVar struct{}

func (groupUser *GroupUser) AddUser(user ...string) {
	for _, s := range user {
		groupUser.usernames[s] = _myVar
	}
}
func (groupUser *GroupUser) GetRawUser() map[string]any {
	return groupUser.usernames
}
