package cluster

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	wf "github.com/chuccp/go-web-frame"
	wfcore "github.com/chuccp/go-web-frame/core"
	wflog "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"go.uber.org/zap"
)

type Service struct {
	app          *core.App
	machineStore *MachineStore
	userStore    *userStore
	grpcSrv      *grpcServer
	ctx          *wfcore.Context
	grpcPort     int
}

func NewService() *Service { return &Service{} }

func (s *Service) Init(ctx *wfcore.Context) error {
	s.app = wf.GetService[*core.App](ctx)
	s.ctx = ctx
	if !s.app.GetCfgBoolDefault("cluster", "start", false) {
		return nil
	}

	grpcClient := NewGrpcClient()
	s.machineStore = NewMachineStore(s.app, grpcClient, ctx)
	s.app.SetForward(s)

	machineId := s.app.GetCfgString("cluster", "machine_id")
	if len(machineId) == 0 {
		machineId = MachineId()
	}
	wflog.Info("machineId", zap.String("machineId", machineId))

	s.grpcPort = s.app.GetCfgInt("cluster", "local_port")
	if s.grpcPort <= 0 {
		s.grpcPort = s.app.GetCfgInt("web", "server.port") + 1
	}

	s.machineStore.localMachine = &Machine{MachineId: machineId, Link: "0.0.0.0:" + strconv.Itoa(s.grpcPort)}

	for _, host := range strings.Split(s.app.GetCfgString("cluster", "remote_host"), ",") {
		host = strings.TrimSpace(host)
		if len(host) == 0 {
			continue
		}
		if !strings.Contains(host, ":") {
			host = "127.0.0.1:" + host
		}
		s.machineStore.addFirstMachine(&Machine{Link: host})
	}

	s.userStore = newUserStore(s.app, s.sendMsg)
	s.app.RegisterHandle("machineInfoId", s.machineInfoId)
	s.app.RegisterHandle("remoteMachineNum", s.remoteMachineNum)
	s.app.RegisterHandle("clusterUserNum", s.clusterUserNum)
	s.app.RegisterHandle("machineAddress", s.machineAddress)

	s.grpcSrv = newGrpcServer(s.app, s.machineStore)
	wflog.Info("gRPC port", zap.Int("port", s.grpcPort))
	return nil
}

func (s *Service) Run() error {
	if s.grpcPort <= 0 {
		return nil
	}

	s.ctx.Go(func(c *wfcore.Context) {
		if err := s.grpcSrv.start(s.grpcPort); err != nil {
			wflog.Error("gRPC start failed", zap.Error(err))
		}
	})
	time.Sleep(time.Second)

	s.ctx.Go(func(c *wfcore.Context) {
		s.loop()
	})
	s.ctx.Go(func(c *wfcore.Context) {
		s.checkUser()
	})

	select {}
}

func (s *Service) loop() {
	for {
		time.Sleep(time.Second * 5)
		s.machineStore.initials()
		time.Sleep(time.Second * 5)
		s.machineStore.queryMachineList()
	}
}

func (s *Service) checkUser() {
	for {
		time.Sleep(time.Second * 5)
		s.userStore.ClearTimeOutUser(time.Now())
		time.Sleep(time.Second * 5)
	}
}

func (s *Service) sendMsg(msg message.IMessage, machineId string) (bool, error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}
	return true, s.machineStore.sendMsg(machineId, data)
}

func (s *Service) WriteSyncMessage(iMessage message.IMessage) (bool, error) {
	switch t := iMessage.(type) {
	case *message.TextMessage:
		username := t.GetString(message.To)
		orderUser := s.userStore.GetOrderUser(username)
		exMachineIds := make([]string, 0)
		for _, u := range orderUser {
			cu := u.(*clientUser)
			fa, _ := cu.WriteSyncMessage(iMessage)
			if fa {
				s.userStore.RefreshUser(username, cu.machineId, s.sendMsg)
				return true, nil
			}
			exMachineIds = append(exMachineIds, cu.machineId)
			s.userStore.DeleteUser(username, cu.machineId)
		}
		machines := s.machineStore.getExMachines(exMachineIds...)
		for _, machine := range machines {
			fa, _ := s.sendMsg(t, machine.MachineId)
			if fa {
				s.userStore.AddUser(username, machine.MachineId, s.sendMsg)
				return true, nil
			}
		}
	}
	return false, core.NoFoundUser
}

func (s *Service) Query(parameter *core.Parameter, localValue any) []any {
	return s.machineStore.Query(parameter, localValue)
}

func (s *Service) machineInfoId(*core.Parameter) any    { return s.machineStore.localMachine.MachineId }
func (s *Service) remoteMachineNum(*core.Parameter) any { return s.machineStore.num() }
func (s *Service) clusterUserNum(*core.Parameter) any   { return s.userStore.Num() }
func (s *Service) machineAddress(parameter *core.Parameter) any {
	id := parameter.GetString("machineId")
	if id == s.machineStore.localMachine.MachineId {
		return s.machineStore.localMachine.Link
	}
	return s.machineStore.GetMachineLink(id)
}
