package cluster

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
)

type Server struct {
	core.IForward
	core.IHttpServer
	context       *core.Context
	localMachine  *Machine
	clientOperate *ClientOperate
	userStore     *userStore
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
	server.clientOperate.run()
}

// 初始化，用于机器之间握手,客户端请求时，返回当前机器信息
func (server *Server) initial(w http.ResponseWriter, re *http.Request) {
	var liteMachine LiteMachine
	err := UnmarshalJsonBody(re, &liteMachine)
	if err == nil {
		server.context.GetLog().Info("接收客户端的握手", zap.String("liteMachine.Link", liteMachine.Link))
		machine, err := parseLiteMachine(&liteMachine, re)
		if err == nil {
			marshal, err := json.Marshal(server.localMachine.getLiteMachine())
			if err == nil {
				server.clientOperate.addNewMachine(machine)
				server.context.GetLog().Info("回馈客户端的握手信息", zap.ByteString("body", marshal))
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
		machine, err := parseLiteMachine(&liteMachine, re)
		if err == nil {
			server.clientOperate.addNewMachine(machine)
			marshal, err := json.Marshal(server.clientOperate.getMachineLite())
			if err == nil {
				w.Write(marshal)
				return
			}
		}
	}
	w.WriteHeader(500)
	w.Write([]byte(err.Error()))
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
			if v == nil {
				w.Write([]byte(""))
				return
			} else {
				marshal, err := json.Marshal(v)
				if err == nil {
					w.Write(marshal)
					return
				}
			}
		}
	}
	w.WriteHeader(500)
}

func (server *Server) HandleAddUser(iUser user.IUser) {
	server.clientOperate.SendAddUser(iUser.GetUsername())
}
func (server *Server) HandleDeleteUser(username string) {
	server.clientOperate.SendDeleteUser(username)
}
func (server *Server) Query(parameter *core.Parameter, localValue any) []any {
	return server.clientOperate.Query(parameter, localValue)
}

func (server *Server) WriteMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	switch t := msg.(type) {
	case *message.TextMessage:
		{
			exMachineId := ""
			un := t.GetString(message.To)
			cu, ok := server.userStore.GetUser(un)
			if ok {
				cl, ok := server.clientOperate.getClient(cu.machineId)
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
			machineId, err := server.clientOperate.sendTextMsg(t, exMachineId)
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
	server.context.GetLog().Info("machineId配置", zap.String("machineId", machineId))
	localLink := server.context.GetCfgString("cluster", "local.link")
	if len(localLink) == 0 {
		localLink = server.GetServerHost()
	}
	localMachine, err := parseLink(localLink)
	if err != nil {
		server.context.GetLog().Panic("解析本地配置localLink失败", zap.Error(err))
		return
	}
	localMachine.MachineId = machineId
	clientOperate := NewClientOperate(server.context, localMachine)
	remoteLinkStr := server.context.GetCfgString("cluster", "remote.link")
	remoteLinks := strings.Split(remoteLinkStr, ",")
	for _, remoteLink := range remoteLinks {
		machine, err := parseLink(remoteLink)
		if err != nil {
			server.context.GetLog().Panic("解析本地配置remoteLink失败", zap.Error(err))
			return
		} else {
			clientOperate.addConfigMachine(machine)
		}
	}
	server.localMachine = localMachine
	server.clientOperate = clientOperate
	server.context.RegisterHandle("machineInfoId", server.machineInfoId)
	server.context.RegisterHandle("remoteMachineNum", server.remoteMachineNum)
	server.context.RegisterHandle("machineAddress", server.machineAddress)
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
		err, fa := server.context.SendNoForwardMessage(&textMessage)
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

func (server *Server) machineAddress(parameter *core.Parameter) any {
	machineId := parameter.GetString("machineId")
	if machineId == server.localMachine.MachineId {
		return server.localMachine.Address + ":" + strconv.Itoa(server.localMachine.Port)
	}
	c, ok := server.clientOperate.getClient(machineId)
	if ok {
		return c.remoteMachine.Address
	}
	return ""
}

func (server *Server) remoteMachineNum(parameter *core.Parameter) any {
	return server.clientOperate.num()
}
