package core

import (
	"github.com/chuccp/httpPush/config"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"log"
	"sync"
)

type Register struct {
	servers *sync.Map
	config  *config.Config
}

func (register *Register) AddServer(server Server) {
	register.servers.LoadOrStore(server.Name(), server)
}
func (register *Register) Create() *HttpPush {
	context := register.getContext()
	return context.GetHttpPush()
}
func (register *Register) getContext() *Context {
	context := newContext(register)
	return context
}
func (register *Register) rangeServer(f func(server Server)) {
	register.servers.Range(func(key, value any) bool {
		f(value.(Server))
		return true
	})
}
func NewRegister(config *config.Config) *Register {
	return &Register{servers: new(sync.Map), config: config}
}

type HttpPush struct {
	context    *Context
	httpServer *util.HttpServer
}

func newHttpPush(context *Context) *HttpPush {
	return &HttpPush{httpServer: util.NewServer(), context: context}
}

func (httpPush *HttpPush) startHttpServer() error {
	port := httpPush.context.GetCfgInt("core", "http.port")
	certFile := httpPush.context.GetCfgString("core", "http.certFile")
	keyFile := httpPush.context.GetCfgString("core", "http.keyFile")
	return httpPush.httpServer.StartAutoTLS(port, certFile, keyFile)
}

const (
	VERSION = "0.1.8"
)

func (httpPush *HttpPush) Start() error {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	httpPush.context.log = logger
	httpPush.context.rangeServer(func(server Server) {
		if s, ok := server.(IHttpServer); ok {
			s.init(httpPush.context)
			go func() {
				err := s.start()
				if err != nil {
					log.Print(err)
				}
			}()
		}
		server.Init(httpPush.context)
		go func() {
			err := server.Start()
			if err != nil {
				log.Print(err)
			}
		}()
	})
	httpPush.context.SetSystemInfo("VERSION", VERSION)
	return httpPush.startHttpServer()
}
