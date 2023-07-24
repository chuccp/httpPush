package cluster

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"io"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	core.IHttpServer
	context      *core.Context
	localMachine *Machine
	clientStore  *ClientStore
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}
func (server *Server) Start() error {
	go server.run()
	return nil
}

func (server *Server) run() {
	server.clientStore.run()
}

// 初始化，用于机器之间握手
func (server *Server) initial(w http.ResponseWriter, re *http.Request) {
	all, err := io.ReadAll(re.Body)
	if err != nil {
		return
	} else {
		var liteMachine LiteMachine
		err = json.Unmarshal(all, &liteMachine)
		if err != nil {
			return
		} else {

			if len(liteMachine.MachineId) == 0 {
				log.Println("提交  MachineId 为空", liteMachine.Link)
			} else {
				machine, err := parseLink(liteMachine.Link)
				if err != nil {
					return
				}
				server.clientStore.addNewMachine(liteMachine.MachineId, machine)
			}
		}
	}
	marshal, err := json.Marshal(server.localMachine.getLiteMachine())
	if err == nil {
		w.Write(marshal)
	}
}

// 查询当前服务器连接的其它机器
func (server *Server) queryMachineList(w http.ResponseWriter, re *http.Request) {
	all, err := io.ReadAll(re.Body)
	if err != nil {
		return
	} else {
		var liteMachine LiteMachine
		err = json.Unmarshal(all, &liteMachine)
		if err != nil {
			return
		} else {
			machine, err := parseLink(liteMachine.Link)
			if err != nil {
				return
			}
			server.clientStore.addNewMachine(liteMachine.MachineId, machine)
		}
	}
	marshal, err := json.Marshal(server.clientStore.getMachineLite())
	if err == nil {
		w.Write(marshal)
	}
}

func (server *Server) Init(context *core.Context) {
	server.context = context
	machineId := server.context.GetCfgString("cluster", "machineId")
	if len(machineId) == 0 {
		machineId = MachineId()
	}
	log.Println("machineId", machineId)
	localLink := server.context.GetCfgString("cluster", "local.link")
	localMachine, err := parseLink(localLink)
	if err != nil {
		log.Panicln(err)
		return
	}
	localMachine.MachineId = machineId
	clientStore := NewClientStore(localMachine)
	remoteLinkStr := server.context.GetCfgString("cluster", "remote.link")
	remoteLinks := strings.Split(remoteLinkStr, ",")
	for _, remoteLink := range remoteLinks {
		machine, err := parseLink(remoteLink)
		if err != nil {
			log.Panicln(err)
			return
		} else {
			clientStore.addMachineNoMachineId(machine)
		}
	}
	server.localMachine = localMachine
	server.clientStore = clientStore
	server.AddHttpRoute("/_cluster/initial", server.initial)
	server.AddHttpRoute("/_cluster/queryMachineList", server.queryMachineList)

}
func (server *Server) Name() string {

	return "cluster"
}
