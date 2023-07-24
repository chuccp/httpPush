package ex

import (
	"github.com/chuccp/httpPush/core"
	"log"
	"net/http"
)

type Server struct {
	core.IHttpServer
	context *core.Context
	store   *Store
}

func NewServer() *Server {
	server := &Server{store: NewStore()}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}

func (server *Server) Start() error {
	server.AddHttpRoute("/ex", server.ex)
	return nil
}
func (server *Server) ex(w http.ResponseWriter, re *http.Request) {
	server.jack(w, re)
}

func (server *Server) jack(w http.ResponseWriter, re *http.Request) {
	cl := NewClient(server.context, re)
	client, ok := server.store.LoadOrStore(cl)
	if ok {
		log.Println("新增用户：", cl.username)
	}
	client.WaitMsg(w, re)
}

func (server *Server) Init(context *core.Context) {
	server.context = context
}
func (server *Server) Name() string {

	return "ex"
}
