package ex

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type Server struct {
	core.IHttpServer
	context  *core.Context
	store    *Store
	liveTime int
	isStart  bool
	rLock    *sync.RWMutex
	tw       *util.TimeWheel
	tw2      *util.TimeWheel2
}

func NewServer() *Server {
	server := &Server{store: NewStore()}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	server.rLock = new(sync.RWMutex)
	server.tw = util.NewTimeWheel(2, 180)
	server.tw2 = util.NewTimeWheel2(1, 60)
	return server
}

func (server *Server) Start() error {
	if server.isStart {
		server.AddHttpRoute("/ex", server.ex)
		server.context.RecoverGo(func() {
			server.tw.Start()
		})
		server.context.RecoverGo(func() {
			server.tw2.Start()
		})
	}
	return nil
}
func (server *Server) ex(w http.ResponseWriter, re *http.Request) {
	util.HttpCross(w)
	server.jack(w, re)
}
func (server *Server) GetTimeWheelLog(parameter *core.Parameter) any {
	return server.tw.GetLog()
}
func (server *Server) jack(writer http.ResponseWriter, re *http.Request) {

	username := util.GetUsername(re)
	if len(username) == 0 {
		writer.WriteHeader(404)
		writer.Write([]byte("request error"))
		return
	}
	server.rLock.RLock()
	cl := getNewClient(server.context, username, server.liveTime)
	_client_, ok := server.store.LoadOrStore(cl, username)
	if ok {
		freeNoUseClient(cl)
	}
	user := _client_.loadUser(writer, re)
	server.tw2.DeleteFunc(user.GetId())
	server.rLock.RUnlock()
	user.waitMessage(server.tw)
	user.RefreshExpired()
	server.tw2.AfterFunc(4, user.GetId(), func(value ...any) {
		server.deleteClientOrUser(value[0].(*client), value[1].(*User))
	}, _client_, user)
}
func (server *Server) deleteClientOrUser(client *client, user *User) {
	server.rLock.Lock()
	if user.expiredTime != nil {
		client.deleteUser(user.GetId())
		if client.userNum() == 0 {
			server.store.Delete(client.username)
			freeClient(client)
		}
		server.context.DeleteUser(user)
	}
	server.rLock.Unlock()
}

func (server *Server) Init(context *core.Context) {
	server.context = context
	server.context.RegisterHandle("getTimeWheelLog", server.GetTimeWheelLog)
	server.isStart = server.context.GetCfgBoolDefault("ex", "start", false)
	if server.isStart {
		server.liveTime = server.context.GetCfgInt("ex", "liveTime")
		server.context.GetLog().Info("ex 配置", zap.Int("liveTime", server.liveTime))
	}
}
func (server *Server) Name() string {

	return "ex"
}
