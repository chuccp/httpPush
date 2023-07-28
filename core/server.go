package core

import (
	"github.com/chuccp/httpPush/util"
	"log"
	"net"
	"net/http"
	"strconv"
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
	GetServerHost() string
}
type httpServer struct {
	context    *Context
	port       int
	usePort    int
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
		server.context.addHttpRoute(pattern, handler)
	}
}
func (server *httpServer) GetServerHost() string {
	if server.IsTls() {
		return "https://" + net.IPv4zero.String() + ":" + strconv.Itoa(server.usePort)
	} else {
		return "http://" + net.IPv4zero.String() + ":" + strconv.Itoa(server.usePort)
	}
}
func (server *httpServer) IsTls() bool {
	if server.port > 0 {
		return server.httpServer.IsTls()
	} else {
		return server.context.isTls()
	}
}

func (server *httpServer) init(context *Context) {
	server.context = context
	port := context.GetCfgInt(server.name, "http.port")
	corePort := context.GetCfgInt("core", "http.port")
	if port > 0 && corePort != port {
		log.Println(server.name, "端口：", port)
		server.certFile = context.GetCfgString(server.name, "http.certFile")
		server.keyFile = context.GetCfgString(server.name, "http.keyFile")
		server.port = port
		server.usePort = port
		server.httpServer = util.NewServer()
	} else {
		server.usePort = corePort
		log.Println(server.name, "端口：", corePort)
	}
}
func (server *httpServer) start() error {
	if server.port > 0 {
		return server.httpServer.StartAutoTLS(server.port, server.certFile, server.keyFile)
	} else {
		return nil
	}
}
