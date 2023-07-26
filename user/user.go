package user

import (
	"github.com/chuccp/httpPush/message"
	"time"
)

type WriteCallBackFunc func(err error, hasUser bool)
type IUser interface {
	WriteMessage(iMessage message.IMessage, writeFunc WriteCallBackFunc)
	GetId() string
	Close()
	GetUsername() string
	GetGroupIds() []string
	GetRemoteAddress() string
	SetUsername(username string)
	LastLiveTime() *time.Time
	CreateTime() *time.Time
}
