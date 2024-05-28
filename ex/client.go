package ex

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"net/http"
	"sync"
	"time"
)

const expiredTime = 5 * time.Second
const defaultExpiredTime = 2 * expiredTime

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
		//c.context.GetLog().Debug("expiredCheck", zap.String("expiredTime", util.FormatTime(u.expiredTime)))
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
func (c *client) userNum() int {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return len(c.connMap)
}
func (c *client) loadUser(writer http.ResponseWriter, re *http.Request) *User {
	liveTime := util.GetLiveTime(re)
	id := getId(c.username, re)
	if liveTime <= 0 {
		if c.liveTime > 0 {
			liveTime = c.liveTime
		} else {
			liveTime = 20
		}
	}
	c.rLock.Lock()
	t := time.Now()
	uv, ok := c.connMap[id]
	u := NewUser(c.username, id, c.queue, c.context, writer, re)
	u.liveTime = liveTime
	u.lastLiveTime = &t
	c.connMap[id] = u
	if !ok {
		u.createTime = &t
		u.addTime = &t
		c.rLock.Unlock()
		c.context.AddUser(u)
		return u
	} else {
		u.createTime = uv.createTime
		u.addTime = uv.addTime
		c.connMap[id] = u
		c.rLock.Unlock()
		return u
	}

}
func getId(username string, re *http.Request) string {
	return username + "_" + re.RemoteAddr
}
