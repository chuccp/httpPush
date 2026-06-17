package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	wfcore "github.com/chuccp/go-web-frame/core"
	wflog "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/httpPush/core"
	ws "github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Controller struct {
	app      *core.App
	store    *Store
	rLock    *sync.RWMutex
	upgrader ws.Upgrader
}

func NewController(app *core.App) *Controller {
	return &Controller{
		app:   app,
		store: NewStore(),
		rLock: new(sync.RWMutex),
		upgrader: ws.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (c *Controller) Init(ctx *wfcore.Context) error {
	if !c.app.GetCfgBoolDefault("ws", "start", false) {
		return nil
	}
	ctx.Any("/ws", c.handleWs)
	return nil
}

func (c *Controller) handleWs(r *web.Request) (any, error) {
	ginCtx := r.GinContext()
	username := ginCtx.Query("id")
	if username == "" {
		username = ginCtx.Query("username")
	}
	if len(username) == 0 {
		return "userId required", nil
	}

	conn, err := c.upgrader.Upgrade(ginCtx.Writer, ginCtx.Request, nil)
	if err != nil {
		return nil, err
	}

	writeCh := make(chan []byte, 16)
	id := username + "_" + ginCtx.Request.RemoteAddr

	c.rLock.RLock()
	cl := getNewClient(c.app, username)
	_client_, ok := c.store.LoadOrStore(cl, username)
	if ok {
		freeClient(cl)
	}
	cuser := NewUser(username, id, c.app, conn, writeCh, ginCtx.Request)
	_client_.connMap.Put(id, cuser)
	c.rLock.RUnlock()

	c.app.AddUser(cuser)
	wflog.Info("ws connect", zap.String("user", username))

	go c.readPump(conn, username)
	c.writePump(conn, writeCh)

	c.rLock.Lock()
	c.app.DeleteUser(cuser)
	cl.connMap.Delete(cuser.GetId())
	if cl.Empty() {
		c.store.Delete(cl.username)
		freeClient(cl)
	}
	c.rLock.Unlock()

	return nil, nil
}

func (c *Controller) writePump(conn *ws.Conn, writeCh chan []byte) {
	ticker := time.NewTicker(30 * time.Second)
	defer func() { ticker.Stop(); conn.Close() }()

	for {
		select {
		case data, ok := <-writeCh:
			if !ok {
				return
			}
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(ws.TextMessage, data); err != nil {
				return
			}
		case <-ticker.C:
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Controller) readPump(conn *ws.Conn, username string) {
	defer conn.Close()
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			wflog.Info("ws disconnect", zap.String("user", username))
			break
		}
		var wsMsg struct {
			To  string `json:"to"`
			Msg string `json:"msg"`
		}
		if json.Unmarshal(msg, &wsMsg) == nil && len(wsMsg.To) > 0 {
			c.app.SendTextMessage(username, wsMsg.To, wsMsg.Msg)
		}
	}
}
