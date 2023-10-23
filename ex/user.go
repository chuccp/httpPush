package ex

import (
	"context"
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type User struct {
	user.IUser
	username      string
	remoteAddress string
	liveTime      int
	writer        http.ResponseWriter
	lastLiveTime  *time.Time
	createTime    *time.Time
	addTime       *time.Time
	last          *time.Time
	queue         *util.Queue
	expiredTime   *time.Time
	groupIds      []string
	context       *core.Context
}
type HttpMessage struct {
	From string
	Body string
}

func newHttpMessage(from string, body string) *HttpMessage {
	return &HttpMessage{From: from, Body: body}
}

func (u *User) GetId() string {
	return u.username + "_" + u.remoteAddress
}

func (u *User) waitMessage() {

	u.context.GetLog().Debug("等待信息", zap.Int("liveTime", u.liveTime))
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(u.liveTime)*time.Second)
	v, num, cls := u.queue.Dequeue(ctx)
	u.context.GetLog().Debug("收到信息：剩余消息", zap.Int32("num", num), zap.Bool("cls", cls))
	if cls {
		u.writer.Write([]byte("[]"))
	} else {
		cancelFunc()
		if v != nil {
			_, err := u.writer.Write(v.([]byte))
			if err != nil {
				u.queue.Offer(v)
			}
		} else {
			u.writer.Write([]byte("[]"))
		}
	}
}

func (u *User) isExpired(now *time.Time) bool {
	if u.expiredTime != nil {
		if u.expiredTime.Before(*now) {
			return true
		}
	}
	return false
}

func (u *User) GetUsername() string {
	return u.username
}

func (u *User) WriteMessage(iMessage message.IMessage, writeFunc user.WriteCallBackFunc) {
	ht := newHttpMessage(
		iMessage.GetString(message.From),
		iMessage.GetString(message.Msg))
	//ht.ExData = iMessage.GetExData()
	hts := []*HttpMessage{ht}
	data, err := json.Marshal(hts)
	if err == nil {
		u.queue.Offer(data)
		writeFunc(nil, true)
	} else {
		writeFunc(err, false)
	}
}

func (u *User) Close() {}

func (u *User) GetGroupIds() []string {
	return u.groupIds
}
func (u *User) GetRemoteAddress() string {
	return u.remoteAddress
}
func (u *User) SetUsername(username string) {
	u.username = username
}
func (u *User) LastLiveTime() *time.Time {
	return u.lastLiveTime
}
func (u *User) CreateTime() *time.Time {
	return u.createTime
}

func NewUser(username string, queue *util.Queue, context *core.Context, writer http.ResponseWriter, re *http.Request) *User {
	u := &User{username: username, context: context, queue: queue, writer: writer, remoteAddress: re.RemoteAddr}
	u.groupIds = util.GetGroupIds(re)
	return u
}
