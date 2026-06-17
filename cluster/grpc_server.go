package cluster

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	wflog "github.com/chuccp/go-web-frame/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
)

type grpcServer struct {
	UnimplementedClusterServer
	ctx          *core.Context
	machineStore *MachineStore
}

func newGrpcServer(ctx *core.Context, store *MachineStore) *grpcServer {
	return &grpcServer{ctx: ctx, machineStore: store}
}

// fixLinkByPeer 使用 peer address 修正机器地址，解决 0.0.0.0/NAT 等问题
func fixLinkByPeer(ctx context.Context, machine *Machine) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return
	}
	peerHost, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return
	}

	// 无 scheme 格式，用 peer IP 修正地址
	if !strings.Contains(machine.Link, "://") {
		_, port, err := net.SplitHostPort(machine.Link)
		if err != nil {
			// 可能只有端口号或只有 IP，直接替换整段
			machine.Link = peerHost + ":" + machine.Link
			return
		}
		machine.Link = peerHost + ":" + port
		return
	}

	// 旧格式 http://host:port，保留 scheme 和 port
	u, err := url.Parse(machine.Link)
	if err != nil {
		return
	}
	machine.Link = u.Scheme + "://" + peerHost + ":" + u.Port()
}

// Initial 握手
func (s *grpcServer) Initial(ctx context.Context, req *InitialRequest) (*InitialResponse, error) {
	var machine Machine
	if err := json.Unmarshal(req.Machine, &machine); err != nil {
		return nil, err
	}
	fixLinkByPeer(ctx, &machine)

	wflog.Info("接收客户端的握手(gRPC)",
		zap.String("machine.Link", machine.Link))

	s.machineStore.addMachine(&machine)

	marshal, err := json.Marshal(s.machineStore.localMachine)
	if err != nil {
		return nil, err
	}
	return &InitialResponse{Machine: marshal}, nil
}

// QueryMachineList 同步节点列表
func (s *grpcServer) QueryMachineList(ctx context.Context, req *QueryMachineListRequest) (*QueryMachineListResponse, error) {
	var machine Machine
	if err := json.Unmarshal(req.Machine, &machine); err != nil {
		return nil, err
	}
	fixLinkByPeer(ctx, &machine)

	wflog.Debug("接收客户端的查询(gRPC)",
		zap.String("machine.Link", machine.Link))

	s.machineStore.addMachine(&machine)

	marshal, err := json.Marshal(s.machineStore.GetMachines())
	if err != nil {
		return nil, err
	}
	return &QueryMachineListResponse{Machines: marshal}, nil
}

// Query 通用查询转发
func (s *grpcServer) Query(ctx context.Context, req *QueryRequest) (*QueryResponse, error) {
	var parameter core.Parameter
	if err := json.Unmarshal(req.Parameter, &parameter); err != nil {
		return nil, err
	}

	handleFunc, fa := s.ctx.GetHandle(parameter.Path)
	if fa {
		v := handleFunc(&parameter)
		if v == nil {
			return &QueryResponse{}, nil
		}
		marshal, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		return &QueryResponse{Result: marshal}, nil
	}
	return &QueryResponse{}, nil
}

// SendTextMsg 转发消息
func (s *grpcServer) SendTextMsg(ctx context.Context, req *SendTextMsgRequest) (*SendTextMsgResponse, error) {
	var textMessage message.TextMessage
	if err := json.Unmarshal(req.Message, &textMessage); err != nil {
		return nil, err
	}
	to := textMessage.GetString(message.To)
	// 群发消息：__grp__:groupId 格式
	if strings.HasPrefix(to, "__grp__:") {
		groupId := strings.TrimPrefix(to, "__grp__:")
		from := textMessage.GetString(message.From)
		body := textMessage.GetString(message.Msg)
		num := s.ctx.SendGroupTextMessage(from, groupId, body)
		wflog.Debug("收到群发消息(gRPC)", zap.String("groupId", groupId), zap.Int32("num", num))
		return &SendTextMsgResponse{Code: 200}, nil
	}
	fa, err := s.ctx.SendLocalMessage(&textMessage)
	wflog.Debug("收到远程信息(gRPC):",
		zap.String("toUser", to),
		zap.Bool("是否成功", fa),
		zap.Error(err))
	if fa {
		return &SendTextMsgResponse{Code: 200}, nil
	}
	if err != nil {
		return &SendTextMsgResponse{Code: 500, Error: err.Error()}, nil
	}
	return &SendTextMsgResponse{Code: 404, Error: "fail"}, nil
}

// startGrpcServer 启动 gRPC 服务
func (s *grpcServer) start(port int) error {
	lis, err := net.Listen("tcp", net.IPv4zero.String()+":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	grpcSrv := grpc.NewServer()
	RegisterClusterServer(grpcSrv, s)
	wflog.Info("gRPC server starting", zap.Int("port", port))
	return grpcSrv.Serve(lis)
}
