package user

import (
	"github.com/chuccp/httpPush/message"
	"sort"
	"time"
)

type WriteCallBackFunc func(err error, hasUser bool)

type IUser interface {
	IOrderUser
	GetId() string
	Close()
	GetUsername() string
	GetGroupIds() []string
	GetRemoteAddress() string
	SetUsername(username string)
	LastLiveTime() *time.Time
	CreateTime() *time.Time
}
type ByAsc []IOrderUser

func (a ByAsc) Len() int      { return len(a) }
func (a ByAsc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByAsc) Less(i, j int) bool {
	if a[i].GetPriority() == a[j].GetPriority() {
		return a[i].GetOrderTime().Compare(*a[j].GetOrderTime()) > 0
	}
	return a[i].GetPriority() < a[j].GetPriority()
}

type IOrderUser interface {
	WriteSyncMessage(iMessage message.IMessage) (bool, error)
	GetPriority() int
	GetMachineId() string
	GetOrderTime() *time.Time
}

func SortByAsc(us []IUser) []IUser {
	ious := make([]IOrderUser, len(us))
	for i, u := range us {
		ious[i] = u.(IOrderUser)
	}
	sort.Sort(ByAsc(ious))
	users := make([]IUser, len(us))
	for i, user := range ious {
		users[i] = user.(IUser)
	}
	return users
}
