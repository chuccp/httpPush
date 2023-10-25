package api

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
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
	if b && err == nil {
		w.Write([]byte("success"))
	} else {
		w.Write([]byte("NO user"))
	}
}
func (server *Server) root(writer http.ResponseWriter, request *http.Request) {
	var dm = make(map[string]interface{})
	dm["version"] = core.VERSION
	data, _ := json.Marshal(dm)
	writer.Write(data)
}
func (server *Server) Start() error {
	server.AddHttpRoute("/sendmsg", server.sendMsg)
	server.AddHttpRoute("/root_version", server.root)
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
