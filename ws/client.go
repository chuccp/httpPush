package ws

import (
	"net/http"
	"sync"

	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
)

type client struct {
	username string
	context  *core.App
	connMap  *util.SliceMap[*User]
	rLock    *sync.RWMutex
}

var poolClient = &sync.Pool{
	New: func() interface{} {
		return &client{connMap: new(util.SliceMap[*User]), rLock: new(sync.RWMutex)}
	},
}

func getNewClient(context *core.App, username string) *client {
	cl := poolClient.Get().(*client)
	cl.connMap.Reset()
	cl.username = username
	cl.context = context
	return cl
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

func getId(username string, re *http.Request) string {
	return username + "_" + re.RemoteAddr
}
