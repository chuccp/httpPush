package ws

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type Server struct {
	core.IHttpServer
	context  *core.Context
	store    *Store
	isStart  bool
	rLock    *sync.RWMutex
	upgrader websocket.Upgrader
}

func NewServer() *Server {
	server := &Server{store: NewStore()}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	server.rLock = new(sync.RWMutex)
	return server
}

func (server *Server) Start() error {
	if server.isStart {
		server.AddHttpRoute("/ws", server.wsHandler)
	}
	return nil
}

func (server *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	username := util.GetUsername(r)
	if len(username) == 0 {
		http.Error(w, "userId required", 400)
		return
	}

	conn, err := server.upgrader.Upgrade(w, r, nil)
	if err != nil {
		server.context.GetLog().Error("ws upgrade failed", zap.Error(err))
		return
	}

	server.rLock.RLock()
	cl := getNewClient(server.context, username)
	_client_, ok := server.store.LoadOrStore(cl, username)
	if ok {
		freeClient(cl)
	}
	id := getId(username, r)
	cuser := NewUser(username, id, server.context, conn, r)
	_client_.connMap.Put(id, cuser)
	server.rLock.RUnlock()

	server.context.AddUser(cuser)
	server.context.GetLog().Info("ws 用户连接", zap.String("username", username), zap.String("remote", r.RemoteAddr))

	// 阻塞读取，直到 WebSocket 断开
	server.readLoop(conn, server.context, username)

	// 清理
	server.deleteClientOrUser(_client_, cuser)
}

func (server *Server) deleteClientOrUser(cl *client, cuser *User) {
	server.rLock.Lock()
	defer server.rLock.Unlock()
	server.context.DeleteUser(cuser)
	cl.connMap.Delete(cuser.GetId())
	if cl.Empty() {
		server.store.Delete(cl.username)
		freeClient(cl)
	}
}

func (server *Server) readLoop(conn *websocket.Conn, context *core.Context, username string) {
	defer conn.Close()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			context.GetLog().Info("ws 用户断开", zap.String("username", username))
			break
		}

		var wsMsg struct {
			To  string `json:"to"`
			Msg string `json:"msg"`
		}
		if json.Unmarshal(msg, &wsMsg) == nil && len(wsMsg.To) > 0 {
			context.SendTextMessage(username, wsMsg.To, wsMsg.Msg)
		}
	}
}

func (server *Server) Init(context *core.Context) {
	server.context = context
	server.isStart = server.context.GetCfgBoolDefault("ws", "start", false)
	if server.isStart {
		server.upgrader = websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		server.context.GetLog().Info("ws 模块启动")
	}
}

func (server *Server) Name() string {
	return "ws"
}
