package cluster

import (
	"github.com/chuccp/httpPush/core"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	core.IHttpServer
	context            *core.Context
	localMachine       *Machine
	remoteMachineStore *MachineStore
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	server.remoteMachineStore = NewMachineStore()
	return server
}
func (server *Server) Start() error {
	go server.run()
	return nil
}

func (server *Server) run() {

}

// 初始化，用于机器之间握手
func (server *Server) initial(w http.ResponseWriter, re *http.Request) {

}

func (server *Server) Init(context *core.Context) {
	server.context = context
	machineId := server.context.GetCfgString("cluster", "machineId")
	if len(machineId) == 0 {
		machineId = MachineId()
	}
	localLink := server.context.GetCfgString("cluster", "local.link")
	machine, err := parseLink(localLink)
	if err != nil {
		log.Panicln(err)
		return
	}
	remoteLinkStr := server.context.GetCfgString("cluster", "remote.link")
	remoteLinks := strings.Split(remoteLinkStr, ",")
	for _, remoteLink := range remoteLinks {
		machine, err := parseLink(remoteLink)
		if err != nil {
			log.Panicln(err)
			return
		} else {
			server.remoteMachineStore.addMachineByLink(machine)
		}
	}
	machine.MachineId = machineId
	server.localMachine = machine
	server.AddHttpRoute("/_cluster/initial", server.initial)

}
func (server *Server) Name() string {

	return "cluster"
}
