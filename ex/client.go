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

var poolClient = &sync.Pool{
	New: func() interface{} {
		connMap := make(map[string]*User)
		queue := util.NewQueue()
		rLock := new(sync.RWMutex)
		return &client{connMap: connMap, queue: queue, rLock: rLock}
	},
}

func getNewClient(context *core.Context, username string, liveTime int) *client {
	client := poolClient.Get().(*client)
	client.username = username
	client.context = context
	client.liveTime = liveTime
	return client
}
func freeNoUseClient(client *client) {
	poolClient.Put(client)
}
func freeClient(client *client) {
	client.connMap = make(map[string]*User)
	client.queue = util.NewQueue()
	client.rLock = new(sync.RWMutex)
	poolClient.Put(client)
}

func (c *client) expiredCheck() {
	c.rLock.Lock()
	t := time.Now()
	keys := make([]string, 0)
	users := make([]*User, 0)
	for key, u := range c.connMap {
		if u.isExpired(&t) {
			keys = append(keys, key)
			users = append(users, u)
		}
	}
	for _, key := range keys {
		delete(c.connMap, key)
	}
	c.rLock.Unlock()
	for _, user := range users {
		c.context.DeleteUser(user)
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
	c.rLock.Lock()
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
		c.rLock.Unlock()
		c.context.AddUser(u)
		return u
	} else {
		uv.liveTime = u.liveTime
		uv.expiredTime = nil
		uv.lastLiveTime = &t
		uv.writer = writer
		c.rLock.Unlock()
		return uv
	}

}
