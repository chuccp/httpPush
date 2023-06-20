package api

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"net/http"
)

type Server struct {
	core.IHttpServer
	context *core.Context
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}

func (server *Server) Start() error {
	server.AddHttpRoute("/root_info", server.systemInfo)
	return nil
}
func (server *Server) systemInfo(w http.ResponseWriter, re *http.Request) {
	var dm = server.context.GetSystemInfo()
	data, _ := json.Marshal(dm)
	w.Write(data)
}

func (server *Server) Init(context *core.Context) {
	server.context = context
}
func (server *Server) Name() string {

	return "api"
}
