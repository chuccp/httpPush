package api

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"net/http"
)

type Server struct {
	context    *core.Context
	port       int
	httpServer *util.HttpServer
	certFile   string
	keyFile    string
}

func (server *Server) Start() error {
	server.addHttpRoute("/root_info", server.systemInfo)
	if server.port > 0 {
		return server.httpServer.StartAutoTLS(server.port, server.certFile, server.keyFile)
	}
	return nil
}
func (server *Server) addHttpRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	if server.port > 0 {
		server.httpServer.AddRoute(pattern, handler)
	} else {
		server.context.AddHttpRoute(pattern, handler)
	}
}
func (server *Server) systemInfo(w http.ResponseWriter, re *http.Request) {
	var dm = server.context.GetSystemInfo()
	data, _ := json.Marshal(dm)
	w.Write(data)
}

func (server *Server) Init(context *core.Context) {
	server.port = -1
	server.context = context
	port := context.GetCfgInt("api", "http.port")
	corePort := context.GetCfgInt("core", "http.port")
	if port > 0 && corePort != port {
		server.certFile = context.GetCfgString("api", "http.certFile")
		server.keyFile = context.GetCfgString("api", "http.keyFile")
		server.port = port
		server.httpServer = util.NewServer()
	}
}
func (server *Server) Name() string {

	return "api"
}
