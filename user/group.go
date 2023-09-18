package user

import "time"

type Group struct {
	createTime   *time.Time
	lastLiveTime *time.Time
}

func NewGroup(user IUser) *Group {
	return &Group{createTime: user.LastLiveTime(), lastLiveTime: user.LastLiveTime()}
}
