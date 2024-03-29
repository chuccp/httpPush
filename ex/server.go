package ex

import (
	"github.com/chuccp/httpPush/core"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	core.IHttpServer
	context  *core.Context
	store    *Store
	liveTime int
	isStart  bool
	rLock    *sync.RWMutex
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
		server.AddHttpRoute("/ex", server.ex)
		go server.expiredCheck()
	}
	return nil
}
func (server *Server) ex(w http.ResponseWriter, re *http.Request) {
	server.jack(w, re)
}

func (server *Server) jack(writer http.ResponseWriter, re *http.Request) {
	cl, err := createClient(server.context, re, server.liveTime)
	if err != nil {
		writer.WriteHeader(404)
		writer.Write([]byte(err.Error()))
		return
	}
	server.rLock.RLock()
	client, ok := server.store.LoadOrStore(cl)
	if !ok {
		server.context.GetLog().Debug("新增连接", zap.String("username", cl.username), zap.String("remoteAddress", re.RemoteAddr))
	}
	user := client.loadUser(writer, re)
	server.rLock.RUnlock()
	user.waitMessage()
	client.setExpired(user)
}

func (server *Server) expiredCheck() {
	for {
		time.Sleep(time.Second * 2)
		server.store.RangeClient(func(c *client) {
			c.expiredCheck()
			server.rLock.Lock()
			if c.userNum() == 0 {
				server.store.Delete(c.username)
			}
			server.rLock.Unlock()
		})
	}
}
func (server *Server) Init(context *core.Context) {
	server.context = context
	server.isStart = server.context.GetCfgBoolDefault("ex", "start", false)
	if server.isStart {
		server.liveTime = server.context.GetCfgInt("ex", "liveTime")
		server.context.GetLog().Info("ex 配置", zap.Int("liveTime", server.liveTime))
	}
}
func (server *Server) Name() string {

	return "ex"
}
