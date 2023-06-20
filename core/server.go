package core

import (
	"github.com/chuccp/httpPush/util"
	"net/http"
)

type Server interface {
	Start() error
	Init(context *Context)
	Name() string
}
type IHttpServer interface {
	AddHttpRoute(pattern string, handler func(http.ResponseWriter, *http.Request))
	init(context *Context)
	start() error
}
type httpServer struct {
	context    *Context
	port       int
	httpServer *util.HttpServer
	certFile   string
	keyFile    string
	name       string
}

func NewHttpServer(name string) IHttpServer {
	return &httpServer{name: name}
}
func (server *httpServer) AddHttpRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if server.port > 0 {
		server.httpServer.AddRoute(pattern, handler)
	} else {
		server.context.AddHttpRoute(pattern, handler)
	}
}
func (server *httpServer) init(context *Context) {
	server.context = context
	port := context.GetCfgInt(server.name, "http.port")
	corePort := context.GetCfgInt("core", "http.port")
	if port > 0 && corePort != port {
		server.certFile = context.GetCfgString(server.name, "http.certFile")
		server.keyFile = context.GetCfgString(server.name, "http.keyFile")
		server.port = port
		server.httpServer = util.NewServer()
	}
}
func (server *httpServer) start() error {
	if server.port > 0 {
		return server.httpServer.StartAutoTLS(server.port, server.certFile, server.keyFile)
	} else {
		return nil
	}
}
