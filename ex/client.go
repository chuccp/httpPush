package ex

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"net/http"
	"sync"
	"time"
)

const expiredCheckTimeSecond = 6

const expiredTime = 4 * time.Second

const defaultLiveTime = 15

type client struct {
	username string
	context  *core.Context
	connMap  *util.SliceMap[*User]
	queue    *util.SliceQueueSafe
	liveTime int
	rLock    *sync.RWMutex
}

var poolClient = &sync.Pool{
	New: func() interface{} {
		return &client{liveTime: defaultLiveTime, connMap: new(util.SliceMap[*User]), queue: util.NewSliceQueueSafe(), rLock: new(sync.RWMutex)}
	},
}

func getNewClient(context *core.Context, username string, liveTime int) *client {
	client := poolClient.Get().(*client)
	client.connMap.Reset()
	client.queue.Reset()
	client.username = username
	client.context = context
	client.liveTime = liveTime
	return client
}

func freeClient(cl *client) {
	poolClient.Put(cl)
}

func (c *client) deleteUser(id string) {
	c.rLock.Lock()
	defer c.rLock.Unlock()
	c.connMap.Delete(id)
}
func (c *client) userNum() int {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return c.connMap.Len()
}
func (c *client) Empty() bool {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return c.connMap.Empty()
}
func (c *client) loadUser(writer http.ResponseWriter, re *http.Request) *User {
	liveTime := util.GetLiveTime(re)
	id := getId(c.username, re)
	if liveTime <= 0 {
		if c.liveTime > 0 {
			liveTime = c.liveTime
		} else {
			liveTime = defaultLiveTime
		}
	}
	c.rLock.Lock()
	t := time.Now()
	uv, ok := c.connMap.Get(id)
	if !ok {
		u := NewUser(c.username, id, c.queue, c.context, writer, re)
		u.liveTime = liveTime
		u.lastLiveTime = &t
		u.createTime = &t
		c.connMap.Put(id, u)
		c.rLock.Unlock()
		c.context.AddUser(u)
		return u
	} else {
		uv.liveTime = liveTime
		uv.writer = writer
		uv.lastLiveTime = &t
		c.rLock.Unlock()
		return uv
	}
}
func getId(username string, re *http.Request) string {
	return username + "_" + re.RemoteAddr
}
