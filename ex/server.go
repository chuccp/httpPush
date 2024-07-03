package ex

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"github.com/panjf2000/ants/v2"
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
	tw       *util.TimeWheel
	waitPool *ants.Pool
}

func NewServer() *Server {
	server := &Server{store: NewStore()}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	server.rLock = new(sync.RWMutex)
	waitPool, err := ants.NewPool(-1)
	if err != nil {
		panic(err)
	}
	server.waitPool = waitPool
	server.tw = util.NewTimeWheel(2, 120)
	return server
}

func (server *Server) Start() error {
	if server.isStart {
		server.AddHttpRoute("/ex", server.ex)
		server.context.Go(func() {
			server.loop()
		})
	}
	return nil
}
func (server *Server) ex(w http.ResponseWriter, re *http.Request) {
	util.HttpCross(w)
	server.jack(w, re)
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
	client, ok := server.store.LoadOrStore(cl, username)
	if ok {
		freeNoUseClient(cl)
	}
	user := client.loadUser(writer, re)
	server.rLock.RUnlock()
	user.waitMessage(server.tw, server.waitPool)
	user.RefreshExpired()
}

func (server *Server) loop() {
	for {
		time.Sleep(time.Second * 1)
		server.store.RangeClient(func(c *client) {
			//过期检查
			c.expiredCheck()
			server.rLock.Lock()
			if c.userNum() == 0 {
				server.store.Delete(c.username)
				freeClient(c)
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
