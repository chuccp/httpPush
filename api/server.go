package api

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"log"
	"net/http"
)

type Server struct {
	core.IHttpServer
	context *core.Context
	query   *Query
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}
func (server *Server) sendMsg(w http.ResponseWriter, re *http.Request) {
	username := util.GetUsername(re)
	msg := util.GetMessage(re)
	if len(username) == 0 || len(msg) == 0 {
		w.WriteHeader(401)
		w.Write([]byte("username or msg can't blank"))
		return
	}
	err, b := server.context.SendTextMessage("system", username, msg)
	log.Println(err, b)
	if b && err == nil {
		w.Write([]byte("success"))
	} else {
		w.Write([]byte("NO user"))
	}
}

func (server *Server) Start() error {
	server.AddHttpRoute("/sendmsg", server.sendMsg)
	server.query.Init()
	return nil
}

func (server *Server) Init(context *core.Context) {
	server.context = context
	server.query = NewQuery(context, server)
}
func (server *Server) Name() string {

	return "api"
}
