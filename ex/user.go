package ex

import (
	"context"
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
	queue         *util.Queue
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

func NewUser(username string, queue *util.Queue, writer http.ResponseWriter, re *http.Request) *User {
	return &User{username: username, queue: queue, writer: writer, remoteAddress: re.RemoteAddr}
}
