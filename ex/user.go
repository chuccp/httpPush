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
	groupIds      []string
}
type HttpMessage struct {
	From   string
	Body   string
	MsgId  uint32
	ExData map[string]string
}

func newHttpMessage(from string, body string, MsgId uint32, exData map[string]string) *HttpMessage {
	return &HttpMessage{From: from, Body: body, MsgId: MsgId, ExData: exData}
}

func (u *User) GetId() string {
	return u.username + "_" + u.remoteAddress
}

func (u *User) waitMessage() {
	log.Println("收到信息：剩余消息:{}===延时:{}", u.liveTime)
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(u.liveTime)*time.Second)
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
	ht := newHttpMessage(
		iMessage.GetString(message.From),
		iMessage.GetString(message.Msg),
		iMessage.GetUint32(message.MId),
		iMessage.GetExData())
	ht.ExData = iMessage.GetExData()
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

func NewUser(username string, queue *util.Queue, writer http.ResponseWriter, re *http.Request) *User {
	u := &User{username: username, queue: queue, writer: writer, remoteAddress: re.RemoteAddr}
	u.groupIds = util.GetGroupIds(re)
	return u
}
