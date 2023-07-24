package user

import (
	"github.com/chuccp/httpPush/message"
	"time"
)

type IUser interface {
	WriteMessage(iMessage message.IMessage) error
	GetId() string
	Close()
	GetUsername() string
	GetGroupIds() []string
	GetRemoteAddress() string
	SetUsername(username string)
	LastLiveTime() *time.Time
	CreateTime() *time.Time
}
