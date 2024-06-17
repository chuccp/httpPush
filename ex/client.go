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

const liveTime = 5 * time.Second

type client struct {
	username string
	context  *core.Context
	connMap  *util.SliceMap[*User]
	queue    *util.Queue
	liveTime int
	rLock    *sync.RWMutex
}

var poolClient = &sync.Pool{
	New: func() interface{} {
		return &client{liveTime: 20}
	},
}

func getNewClient(context *core.Context, username string, liveTime int) *client {
	client := poolClient.Get().(*client)
	if client.liveTime > 0 {
		client.connMap = GetSliceMap()
		client.queue = util.GetQueue()
		client.rLock = new(sync.RWMutex)
	}
	client.username = username
	client.context = context
	client.liveTime = liveTime
	return client
}
func freeNoUseClient(client *client) {
	client.liveTime = -1
	poolClient.Put(client)
}
func freeClient(client *client) {
	FreeSliceMap(client.connMap)
	util.FreeQueue(client.queue)
	client.liveTime = 20
	client.rLock = nil
	client.queue = nil
	client.connMap = nil
	poolClient.Put(client)
}

func (c *client) expiredCheck() {
	c.rLock.Lock()
	t := time.Now()
	keys := make([]string, 0)
	users := make([]*User, 0)

	c.connMap.Each(func(key string, u *User) {
		if u.isExpired(&t) {
			keys = append(keys, key)
			users = append(users, u)
		}
	})
	for _, key := range keys {
		c.connMap.Delete(key)
	}
	c.rLock.Unlock()
	for _, user := range users {
		c.context.DeleteUser(user)
	}
}
func (c *client) writeCheck() {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	t := time.Now()
	c.connMap.Each(func(key string, u *User) {
		if u.isWriteLive(&t) {
			u.writeLive()
		}
	})
}
func (c *client) userNum() int {
	c.rLock.RLock()
	defer c.rLock.RUnlock()
	return c.connMap.Len()
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
	uv, ok := c.connMap.Get(id)
	if !ok {
		u := NewUser(c.username, id, c.queue, c.context, writer, re)
		u.expiredTime = nil
		u.liveTime = liveTime
		u.writeLiveTime = nil
		u.lastLiveTime = &t
		u.createTime = &t
		u.addTime = &t
		c.connMap.Put(id, u)
		c.rLock.Unlock()
		c.context.AddUser(u)
		return u
	} else {
		uv.liveTime = liveTime
		uv.expiredTime = nil
		uv.lastLiveTime = &t
		uv.writeLiveTime = nil
		uv.writer = writer
		c.rLock.Unlock()
		return uv
	}

}
func getId(username string, re *http.Request) string {
	return username + "_" + re.RemoteAddr
}

var poolSliceMap = &sync.Pool{
	New: func() interface{} {
		return new(util.SliceMap[*User])
	},
}

func GetSliceMap() *util.SliceMap[*User] {
	sliceMap := poolSliceMap.Get().(*util.SliceMap[*User])
	sliceMap.Reset()
	return sliceMap
}
func FreeSliceMap(sliceMap *util.SliceMap[*User]) {
	poolSliceMap.Put(sliceMap)
}
