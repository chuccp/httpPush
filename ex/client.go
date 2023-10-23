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
	connMap  map[string]*User
	queue    *util.Queue
	liveTime int
	rLock    *sync.RWMutex
}

func NewClient(context *core.Context, re *http.Request, liveTime int) *client {
	username := util.GetUsername(re)
	connMap := make(map[string]*User)
	queue := util.NewQueue()
	return &client{queue: queue, username: username, context: context, connMap: connMap, liveTime: liveTime, rLock: new(sync.RWMutex)}
}
func (c *client) expiredCheck() {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	t := time.Now()
	keys := make([]string, 0)
	for key, u := range c.connMap {
		if u.isExpired(&t) {
			keys = append(keys, key)
			c.context.DeleteUser(u)
		}
	}
	for _, key := range keys {
		delete(c.connMap, key)
	}
}
func (c *client) setExpired(user *User) {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	t := time.Now()
	user.last = &t
	tm := t.Add(5 * time.Second)
	user.expiredTime = &tm
}
func (c *client) userNum() int {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return len(c.connMap)
}
func (c *client) loadUser(writer http.ResponseWriter, re *http.Request) *User {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	t := time.Now()
	liveTime := util.GetLiveTime(re)
	u := NewUser(c.username, c.queue, c.context, writer, re)
	if liveTime > 0 {
		u.liveTime = liveTime
	} else if c.liveTime > 0 {
		u.liveTime = c.liveTime
	} else {
		u.liveTime = 20
	}
	u.expiredTime = nil
	id := u.GetId()
	uv, ok := c.connMap[id]
	if !ok {
		c.connMap[id] = u
		u.lastLiveTime = &t
		u.createTime = &t
		u.addTime = &t
		c.context.AddUser(u)
		return u
	} else {
		uv.liveTime = u.liveTime
		uv.expiredTime = nil
		uv.lastLiveTime = &t
		uv.writer = writer
		return uv
	}

}
