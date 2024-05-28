package ex

import (
	"context"
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"net/http"
	"time"
)

type User struct {
	user.IUser
	username      string
	remoteAddress string
	liveTime      int
	priority      int
	writer        http.ResponseWriter
	lastLiveTime  *time.Time
	createTime    *time.Time
	addTime       *time.Time
	last          *time.Time
	queue         *util.Queue
	expiredTime   *time.Time
	groupIds      []string
	context       *core.Context
	id            string
}
type HttpMessage struct {
	From string
	Body string
}

func newHttpMessage(from string, body string) *HttpMessage {
	return &HttpMessage{From: from, Body: body}
}

func (u *User) GetId() string {
	return u.id
}

func messageToBytes(iMessage message.IMessage) ([]byte, error) {
	ht := newHttpMessage(
		iMessage.GetString(message.From),
		iMessage.GetString(message.Msg))
	hts := []*HttpMessage{ht}
	data, err := json.Marshal(hts)
	return data, err
}

func (u *User) RefreshExpired() {
	t := time.Now()
	u.last = &t
	tm := t.Add(expiredTime)
	u.expiredTime = &tm
}

func (u *User) waitMessage() {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(u.liveTime)*time.Second)
	msg, _, cls := u.queue.Dequeue(ctx)
	if cls {
		u.writer.Write([]byte("[]"))
	} else {
		cancelFunc()
		mg, ok := (msg).(message.IMessage)
		if ok {
			v, err := messageToBytes(mg)
			if err == nil && v != nil {
				_, err := u.writer.Write(v)
				if err != nil {
					u.queue.Offer(v)
				} else {
					u.context.RecordMessage(mg)
				}
			} else {
				u.writer.Write([]byte("[]"))
			}
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
	u.queue.Offer(iMessage)
	writeFunc(nil, true)
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

func (u *User) GetPriority() int {
	return u.priority
}
func (u *User) GetMachineId() string {
	return ""
}
func (u *User) GetOrderTime() *time.Time {
	return u.lastLiveTime
}
func NewUser(username string, id string, queue *util.Queue, context *core.Context, writer http.ResponseWriter, re *http.Request) *User {
	u := &User{username: username, id: id, context: context, queue: queue, writer: writer, remoteAddress: re.RemoteAddr}
	u.groupIds = util.GetGroupIds(re)
	return u
}
