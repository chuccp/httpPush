package cluster

import (
	"github.com/chuccp/httpPush/core"
	"log"
	"net/http"
)

type Server struct {
	core.IHttpServer
	context            *core.Context
	localMachine       *Machine
	tempMachineStore   *MachineStore
	remoteMachineStore *MachineStore
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	server.remoteMachineStore = NewMachineStore()
	server.tempMachineStore = NewMachineStore()
	return server
}
func (server *Server) Start() error {

	return nil
}

func (server *Server) basicInfo(w http.ResponseWriter, re *http.Request) {

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
	machine.MachineId = machineId
	server.localMachine = machine
	server.AddHttpRoute("/_cluster/basicInfo", server.basicInfo)
}
func (server *Server) Name() string {

	return "cluster"
}
