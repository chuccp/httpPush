package cluster

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/chuccp/httpPush/core"
	"go.uber.org/zap"
	wflog "github.com/chuccp/go-web-frame/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GrpcClient gRPC 连接管理器，每个远程节点一条常驻连接（HTTP/2 多路复用）
type GrpcClient struct {
	conns map[string]*grpc.ClientConn
	lock  *sync.RWMutex
}

func NewGrpcClient() *GrpcClient {
	return &GrpcClient{
		conns: make(map[string]*grpc.ClientConn),
		lock:  new(sync.RWMutex),
	}
}

func (c *GrpcClient) getConn(link string) (*grpc.ClientConn, error) {
	c.lock.RLock()
	conn, ok := c.conns[link]
	c.lock.RUnlock()
	if ok {
		return conn, nil
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	conn, ok = c.conns[link]
	if ok {
		return conn, nil
	}

	// 去掉 http:// 或 https:// 前缀，gRPC 使用纯 host:port 格式
	grpcTarget := link
	grpcTarget = strings.TrimPrefix(grpcTarget, "https://")
	grpcTarget = strings.TrimPrefix(grpcTarget, "http://")

	conn, err := grpc.NewClient(grpcTarget,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: time.Second,
		}),
	)
	if err != nil {
		return nil, err
	}
	c.conns[link] = conn
	wflog.Info("创建 gRPC 连接", zap.String("link", link))
	return conn, nil
}

// Call 根据 path 路由到对应的 gRPC 方法
func (c *GrpcClient) Call(machine *Machine, path string, jsonData []byte) ([]byte, error) {
	return c.CallByLink(machine.Link, path, jsonData)
}

// CallByLink 根据 path 路由到对应的 gRPC 方法
func (c *GrpcClient) CallByLink(link string, path string, jsonData []byte) ([]byte, error) {
	conn, err := c.getConn(link)
	if err != nil {
		return nil, err
	}
	client := NewClusterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	switch path {
	case "/_cluster/initial":
		resp, err := client.Initial(ctx, &InitialRequest{Machine: jsonData})
		if err != nil {
			return nil, err
		}
		return resp.Machine, nil

	case "/_cluster/queryMachineList":
		resp, err := client.QueryMachineList(ctx, &QueryMachineListRequest{Machine: jsonData})
		if err != nil {
			return nil, err
		}
		return resp.Machines, nil

	case "/_cluster/query":
		resp, err := client.Query(ctx, &QueryRequest{Parameter: jsonData})
		if err != nil {
			return nil, err
		}
		return resp.Result, nil

	case "/_cluster/sendTextMsg":
		resp, err := client.SendTextMsg(ctx, &SendTextMsgRequest{Message: jsonData})
		if err != nil {
			return nil, err
		}
		if resp.Code == 200 {
			return []byte(`{"code":200}`), nil
		}
		if resp.Code == 404 {
			return nil, core.NoFoundUser
		}
		return nil, errors.New(resp.Error)

	default:
		return nil, errors.New("unknown path: " + path)
	}
}
