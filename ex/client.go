package ex

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"net/http"
	"sync"
	"time"
)

type client struct {
	username string
	context  *core.Context
	connMap  *sync.Map
	queue    *util.Queue
}

func NewClient(context *core.Context, re *http.Request) *client {
	username := util.GetUsername(re)
	connMap := new(sync.Map)
	queue := util.NewQueue()
	return &client{queue: queue, username: username, context: context, connMap: connMap}
}
func (c *client) WaitMsg(writer http.ResponseWriter, re *http.Request) {
	user := c.loadUser(writer, re)
	user.waitMessage()
	t := time.Now()
	user.last = &t
	tm := t.Add(time.Duration(user.liveTime) * time.Second)
	user.expiredTime = &tm
}
func (c *client) expiredCheck() {
	t := time.Now()
	c.connMap.Range(func(key, value any) bool {
		u := value.(*User)
		if u.isExpired(&t) {
			id := u.GetId()
			c.connMap.Delete(id)
			c.context.DeleteUser(u)
		}
		return true
	})
}

func (c *client) loadUser(writer http.ResponseWriter, re *http.Request) *User {
	t := time.Now()
	liveTime := util.GetLiveTime(re)
	u := NewUser(c.username, c.queue, writer, re)
	u.liveTime = liveTime
	id := u.GetId()
	v, ok := c.connMap.LoadOrStore(id, u)
	if !ok {
		u.lastLiveTime = &t
		u.createTime = &t
		u.addTime = &t
		c.context.AddUser(u)
		return u
	} else {
		uv := v.(*User)
		uv.expiredTime = nil
		uv.lastLiveTime = &t
		uv.writer = writer
		return uv
	}

}
