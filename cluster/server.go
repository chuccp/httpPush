package cluster

import (
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
	server.AddHttpRoute("/_cluster/basicInfo", server.basicInfo)
	return nil
}

func (server *Server) basicInfo(w http.ResponseWriter, re *http.Request) {

}

func (server *Server) Init(context *core.Context) {
	server.context = context
}
func (server *Server) Name() string {

	return "cluster"
}
