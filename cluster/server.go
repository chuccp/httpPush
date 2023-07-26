package cluster

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"log"
	"net/http"
	"strings"
)

type Server struct {
	core.IForward
	core.IHttpServer
	context      *core.Context
	localMachine *Machine
	clientStore  *ClientStore
	userStore    *userStore
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	server.userStore = newUserStore()
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
	var liteMachine LiteMachine
	err := UnmarshalJsonBody(re, &liteMachine)
	if err == nil {
		machine, err := parseLink(liteMachine.Link)
		if err == nil {
			server.clientStore.addNewMachine(liteMachine.MachineId, machine)
			marshal, err := json.Marshal(server.localMachine.getLiteMachine())
			if err == nil {
				w.Write(marshal)
				return
			}
		}
	}
	w.WriteHeader(500)
}

// 查询当前服务器连接的其它机器
func (server *Server) queryMachineList(w http.ResponseWriter, re *http.Request) {
	var liteMachine LiteMachine
	err := UnmarshalJsonBody(re, &liteMachine)
	if err == nil {
		machine, err := parseLink(liteMachine.Link)
		if err == nil {
			server.clientStore.addNewMachine(liteMachine.MachineId, machine)
			marshal, err := json.Marshal(server.clientStore.getMachineLite())
			if err == nil {
				w.Write(marshal)
				return
			}
		}
	}
	w.WriteHeader(500)
}

func (server *Server) query(w http.ResponseWriter, re *http.Request) {
	var parameter core.Parameter
	err := UnmarshalJsonBody(re, &parameter)
	if err != nil {
		return
	} else {
		handleFunc, fa := server.context.GetHandle(parameter.Path)
		if fa {
			v := handleFunc(&parameter)
			marshal, err := json.Marshal(v)
			if err == nil {
				w.Write(marshal)
				return
			}
		}
	}

	w.WriteHeader(500)
}

func (server *Server) HandleAddUser(iUser user.IUser) {
	server.clientStore.SendAddUser(iUser.GetUsername())
}
func (server *Server) HandleDeleteUser(username string) {
	server.clientStore.SendDeleteUser(username)
}
func (server *Server) Query(parameter *core.Parameter, localValue any) []any {
	return server.clientStore.Query(parameter, localValue)
}

func (server *Server) WriteMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	log.Println("Server!!!!!!!")
	switch t := msg.(type) {
	case *message.TextMessage:
		{
			exMachineId := ""
			un := t.GetString(message.To)
			cu, ok := server.userStore.GetUser(un)
			if ok {
				cl, ok := server.clientStore.getClient(cu.machineId)
				if ok {
					err := cl.sendTextMsg(t)
					if err == nil {
						writeFunc(nil, true)
						return
					} else {
						exMachineId = cl.remoteMachine.MachineId
					}
				}
			}
			machineId, err := server.clientStore.sendTextMsg(t, exMachineId)
			if err == nil {
				server.userStore.AddUser(un, machineId)
				writeFunc(nil, true)
				return
			} else {
				writeFunc(err, false)
				return
			}
		}
	}
	writeFunc(nil, false)
}

func (server *Server) Init(context *core.Context) {
	server.context = context
	context.SetForward(server)
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
	clientStore := NewClientStore(server.context, localMachine)
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
	server.context.RegisterHandle("machineInfoId", server.machineInfoId)
	server.AddHttpRoute("/_cluster/initial", server.initial)
	server.AddHttpRoute("/_cluster/deleteUser", server.deleteUser)
	server.AddHttpRoute("/_cluster/addUser", server.addUser)
	server.AddHttpRoute("/_cluster/queryMachineList", server.queryMachineList)
	server.AddHttpRoute("/_cluster/query", server.query)
	server.AddHttpRoute("/_cluster/sendTextMsg", server.sendTextMsg)
}
func (server *Server) Name() string {

	return "cluster"
}

func (server *Server) deleteUser(writer http.ResponseWriter, request *http.Request) {
	var us []*User
	err := UnmarshalJsonBody(request, &us)
	if err == nil {
		for _, u := range us {
			server.userStore.DeleteUser(u.UserId)
		}
	}
}
func (server *Server) addUser(writer http.ResponseWriter, request *http.Request) {
	var us []*User
	err := UnmarshalJsonBody(request, &us)
	if err == nil {
		for _, u := range us {
			server.userStore.AddUser(u.UserId, u.MachineId)
		}
	}
}

func (server *Server) sendTextMsg(writer http.ResponseWriter, request *http.Request) {
	var textMessage message.TextMessage
	err := UnmarshalJsonBody(request, &textMessage)
	if err == nil {
		err, fa := server.context.SendMessage(&textMessage)
		if fa {
			v, err := json.Marshal(successResponse())
			if err == nil {
				writer.Write(v)
				return
			}
		} else {
			if err != nil {
				v, err := json.Marshal(failResponse(err.Error()))
				if err == nil {
					writer.Write(v)
					return
				}
			} else {
				v, err := json.Marshal(failResponse("fail"))
				if err == nil {
					writer.Write(v)
					return
				}
			}
		}
	}
	writer.WriteHeader(500)
}

func (server *Server) machineInfoId(parameter *core.Parameter) any {
	return server.localMachine.MachineId
}