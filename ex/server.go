package ex

import (
	"sync"
	"time"

	wf "github.com/chuccp/go-web-frame"
	wfcore "github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
)

type Controller struct {
	app      *core.App
	store    *Store
	liveTime int
	rLock    *sync.RWMutex
	tw       *util.TimeWheel2
	tw2      *util.TimeWheel2
}

func NewController() *Controller {
	return &Controller{
		store: NewStore(),
		rLock: new(sync.RWMutex),
		tw:    util.NewTimeWheel2(1, 180),
		tw2:   util.NewTimeWheel2(1, 60),
	}
}

func (c *Controller) Init(ctx *wfcore.Context) error {
	c.app = wf.GetService[*core.App](ctx)
	if !c.app.GetCfgBoolDefault("ex", "start", false) {
		return nil
	}
	c.liveTime = c.app.GetCfgInt("ex", "live_time")

	go c.tw.Start()
	go c.tw2.Start()

	ctx.Any("/ex", c.handleEx)
	return nil
}

func (c *Controller) handleEx(r *web.Request) (any, error) {
	ginCtx := r.GinContext()
	username := util.GetUsername(ginCtx.Request)
	if len(username) == 0 {
		ginCtx.String(404, "request error")
		return nil, nil
	}

	c.rLock.RLock()
	cl := getNewClient(c.app, username, c.liveTime)
	_client_, ok := c.store.LoadOrStore(cl, username)
	if ok {
		freeClient(cl)
	}
	u := _client_.loadUser(ginCtx.Writer, ginCtx.Request)
	u.RefreshPreExpired()
	c.rLock.RUnlock()

	u.waitMessage(c.tw)
	c.tw2.AfterFunc(expiredCheckTimeSecond, u.GetId(), func(value ...any) {
		uu := value[1].(*User)
		cc := value[0].(*client)
		c.deleteClientOrUser(cc, uu)
	}, _client_, u)
	u.RefreshExpired()

	return nil, nil
}

func (c *Controller) deleteClientOrUser(cl *client, u *User) {
	c.rLock.Lock()
	if u.expiredTime != nil && u.expiredTime.Before(time.Now()) {
		u.expiredTime = nil
		c.app.DeleteUser(u)
		cl.deleteUser(u.GetId())
		if cl.Empty() {
			c.store.Delete(cl.username)
			freeClient(cl)
		}
	}
	c.rLock.Unlock()
}
