package cluster

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"go.uber.org/zap"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Server struct {
	core.IForward
	core.IHttpServer
	context      *core.Context
	isStart      bool
	machineStore *MachineStore
	userStore    *userStore
}

func NewServer() *Server {
	server := &Server{}
	httpServer := core.NewHttpServer(server.Name())
	server.IHttpServer = httpServer
	return server
}
func (server *Server) checkUser() {
	for {
		time.Sleep(time.Second * 5)
		server.userStore.ClearTimeOutUser(time.Now())
		time.Sleep(time.Second * 5)
	}
}

func (server *Server) Start() error {
	if server.isStart {
		server.context.RecoverGo(func() {
			server.loop()
		})
		server.context.RecoverGo(func() {
			server.checkUser()
		})
	}
	return nil
}

func (server *Server) loop() {
	for {
		time.Sleep(time.Second * 5)
		server.machineStore.initials()
		time.Sleep(time.Second * 5)
		server.machineStore.queryMachineList()
	}
}

func (server *Server) sendMsg(message message.IMessage, machineId string) (bool, error) {
	marshal, err := json.Marshal(message)
	if err != nil {
		return false, err
	}
	err = server.machineStore.sendMsg(machineId, marshal)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (server *Server) WriteSyncMessage(iMessage message.IMessage) (fa bool, err error) {
	switch t := iMessage.(type) {
	case *message.TextMessage:
		{
			username := t.GetString(message.To)
			orderUser := server.userStore.GetOrderUser(username)
			exMachineIds := make([]string, 0)
			for _, iOrderUser := range orderUser {
				cu := iOrderUser.(*clientUser)
				fa, err = cu.WriteSyncMessage(iMessage)
				server.context.GetLog().Debug("WriteSyncMessage", zap.Bool("fa", fa), zap.Error(err))
				if fa {
					server.userStore.RefreshUser(username, cu.machineId, server.sendMsg)
					return
				} else {
					exMachineIds = append(exMachineIds, cu.machineId)
					server.userStore.DeleteUser(username, cu.machineId)
				}
			}
			machines := server.machineStore.getExMachines(exMachineIds...)
			if len(machines) > 0 {
				for _, machine := range machines {
					fa, err = server.sendMsg(t, machine.MachineId)
					if fa {
						return
					}
				}
			} else {
				return
			}
		}
	}

	return false, core.NoFoundUser
}
func (server *Server) Query(parameter *core.Parameter, localValue any) []any {
	return server.machineStore.Query(parameter, localValue)
}

func (server *Server) machineInfoId(parameter *core.Parameter) any {
	return server.machineStore.localMachine.MachineId
}
func (server *Server) remoteMachineNum(parameter *core.Parameter) any {
	return server.machineStore.num()
}

func (server *Server) initial(w http.ResponseWriter, re *http.Request) {
	machine, err := getRemoteMachine(re)
	if err == nil {
		server.context.GetLog().Info("接收客户端的握手", zap.String("machine.Link", machine.Link), zap.String("remoteAddress", re.RemoteAddr))
		server.machineStore.addMachine(machine)
		marshal, err := json.Marshal(server.machineStore.localMachine)
		if err == nil {
			w.Write(marshal)
		} else {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}

func (server *Server) queryMachineList(w http.ResponseWriter, re *http.Request) {
	machine, err := getRemoteMachine(re)
	if err == nil {
		server.context.GetLog().Debug("接收客户端的查询", zap.String("machine.Link", machine.Link), zap.String("remoteAddress", re.RemoteAddr))
		server.machineStore.addMachine(machine)
		marshal, err := json.Marshal(server.machineStore.GetMachines())
		if err == nil {
			w.Write(marshal)
		} else {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
		}
	} else {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
	}
}

func getRemoteMachine(re *http.Request) (*Machine, error) {
	all, err := io.ReadAll(re.Body)
	if err != nil {
		return nil, err
	} else {
		var machine Machine
		err := json.Unmarshal(all, &machine)
		if err != nil {
			return nil, err
		}
		url, err := url.Parse(machine.Link)
		if err != nil {
			return nil, err
		}
		host, _, err := net.SplitHostPort(re.RemoteAddr)
		if err != nil {
			return nil, err
		}
		link := url.Scheme + "://" + host + ":" + url.Port()
		machine.Link = link
		return &machine, nil
	}
}

func (server *Server) query(w http.ResponseWriter, re *http.Request) {
	var parameter core.Parameter
	err := UnmarshalJsonBody(re, &parameter)
	server.context.GetLog().Debug("收到查询", zap.String("remoteAddress", re.RemoteAddr))
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
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
				} else {
					w.WriteHeader(500)
					w.Write([]byte(err.Error()))
				}
			}
		}
	}

}

func (server *Server) sendTextMsg(writer http.ResponseWriter, request *http.Request) {
	var textMessage message.TextMessage
	err := UnmarshalJsonBody(request, &textMessage)
	if err == nil {
		err, fa := server.context.SendLocalMessage(&textMessage)
		server.context.GetLog().Debug("收到远程信息:", zap.String("toUser", textMessage.GetString(message.To)), zap.Bool("是否成功", fa), zap.Error(err))
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

func (server *Server) Init(context *core.Context) {
	server.context = context
	server.isStart = server.context.GetCfgBoolDefault("cluster", "start", false)
	if server.isStart {
		server.machineStore = NewMachineStore(server.context)
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
		server.machineStore.localMachine = localMachine

		remoteLinkStr := server.context.GetCfgString("cluster", "remote.link")
		remoteLinks := strings.Split(remoteLinkStr, ",")
		for _, remoteLink := range remoteLinks {
			machine, err := parseLink(remoteLink)
			if err != nil {
				server.context.GetLog().Panic("解析本地配置remoteLink失败", zap.Error(err))
				return
			} else {
				server.machineStore.addFirstMachine(machine)
			}
		}
		server.userStore = newUserStore(context, server.sendMsg)
		server.context.RegisterHandle("machineInfoId", server.machineInfoId)
		server.context.RegisterHandle("remoteMachineNum", server.remoteMachineNum)
		server.context.RegisterHandle("clusterUserNum", server.clusterUserNum)
		server.context.RegisterHandle("machineAddress", server.machineAddress)
		server.AddHttpRoute("/_cluster/initial", server.initial)
		server.AddHttpRoute("/_cluster/queryMachineList", server.queryMachineList)
		server.AddHttpRoute("/_cluster/query", server.query)
		server.AddHttpRoute("/_cluster/sendTextMsg", server.sendTextMsg)
	}

}
func (server *Server) Name() string {
	return "cluster"
}

func (server *Server) clusterUserNum(parameter *core.Parameter) any {
	return server.userStore.Num()
}

func (server *Server) machineAddress(parameter *core.Parameter) any {
	machineId := parameter.GetString("machineId")
	if machineId == server.machineStore.localMachine.MachineId {
		return server.machineStore.localMachine.Link
	}
	return server.machineStore.GetMachineLink(machineId)
}
