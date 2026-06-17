package cluster

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"go.uber.org/zap"
)

type Server struct {
	core.IForward
	core.IHttpServer
	context      *core.Context
	isStart      bool
	machineStore *MachineStore
	userStore    *userStore
	grpcServer   *grpcServer
	grpcPort     int
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
		if server.grpcServer != nil && server.grpcPort > 0 {
			server.context.RecoverGo(func() {
				err := server.grpcServer.start(server.grpcPort)
				if err != nil {
					server.context.GetLog().Error("gRPC server 启动失败", zap.Error(err))
				}
			})
		}
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
						server.userStore.AddUser(username, machine.MachineId, server.sendMsg)
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

func (server *Server) Init(context *core.Context) {
	server.context = context
	server.isStart = server.context.GetCfgBoolDefault("cluster", "start", false)
	if server.isStart {
		grpcClient := NewGrpcClient(server.context.GetLog())
		server.machineStore = NewMachineStore(server.context, grpcClient)
		context.SetForward(server)

		machineId := server.context.GetCfgString("cluster", "machineId")
		if len(machineId) == 0 {
			machineId = MachineId()
		}
		server.context.GetLog().Info("machineId配置", zap.String("machineId", machineId))

		// gRPC 端口：cluster.local.port，默认 HTTP端口+1
		httpPort := server.context.GetCfgInt("core", "http.port")
		server.grpcPort = server.context.GetCfgInt("cluster", "local.port")
		if server.grpcPort <= 0 || server.grpcPort == httpPort {
			server.grpcPort = httpPort + 1
		}

		// 本地机器信息，link 使用 gRPC 端口
		localMachine := &Machine{MachineId: machineId, Link: "0.0.0.0:" + strconv.Itoa(server.grpcPort)}
		server.machineStore.localMachine = localMachine

		// 远程节点
		remoteHostStr := server.context.GetCfgString("cluster", "remote.host")
		for _, host := range strings.Split(remoteHostStr, ",") {
			host = strings.TrimSpace(host)
			if len(host) > 0 {
				server.machineStore.addFirstMachine(&Machine{Link: host})
			}
		}

		server.userStore = newUserStore(context, server.sendMsg)
		server.context.RegisterHandle("machineInfoId", server.machineInfoId)
		server.context.RegisterHandle("remoteMachineNum", server.remoteMachineNum)
		server.context.RegisterHandle("clusterUserNum", server.clusterUserNum)
		server.context.RegisterHandle("machineAddress", server.machineAddress)

		// 创建 gRPC server
		server.grpcServer = newGrpcServer(server.context, server.machineStore)
		server.context.GetLog().Info("gRPC 端口配置", zap.Int("grpcPort", server.grpcPort))
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
