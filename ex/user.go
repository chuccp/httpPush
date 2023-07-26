package ex

import (
	"context"
	"encoding/json"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"log"
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
	waitTime := time.Minute
	if u.liveTime > 0 {
		waitTime = 2 * time.Duration(u.liveTime) * time.Second
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), waitTime)
	v, num, cls := u.queue.Dequeue(ctx)
	log.Println("收到信息：剩余消息:{}===延时:{}", num, cls)
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
	ht := newHttpMessage(iMessage.GetString(message.From), iMessage.GetString(message.To))
	hts := []*HttpMessage{ht}
	data, err := json.Marshal(hts)
	if err == nil {
		u.queue.Offer(data)
	} else {
		writeFunc(err, false)
	}
	writeFunc(nil, true)
}

func (u *User) Close() {}

func (u *User) GetGroupIds() []string {
	return []string{}
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

func NewUser(username string, queue *util.Queue, writer http.ResponseWriter, re *http.Request) *User {
	return &User{username: username, queue: queue, writer: writer, remoteAddress: re.RemoteAddr}
}
