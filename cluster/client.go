package cluster

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"sync"
	"time"
)

type client struct {
	remoteMachine *Machine
	localMachine  *Machine
	request       *util.Request
	isHandshake   bool
	isLocal       bool
	remoteLink    string
	context       *core.Context
}

type operate struct {
	isAdd    bool
	username string
}

func NewClient(remoteMachine *Machine, localMachine *Machine, context *core.Context) *client {

	return &client{request: util.NewRequest(), remoteLink: remoteMachine.Link, isHandshake: false, remoteMachine: remoteMachine, localMachine: localMachine, context: context}
}
func (client *client) run() {
	err := client.initial()
	if err == nil {
		client.context.GetLog().Info("握手", zap.String("client.remoteMachine.Link", client.remoteMachine.Link), zap.String("query", "/_cluster/initial"))
	} else {
		client.context.GetLog().Info("握手", zap.String("client.remoteMachine.Link", client.remoteMachine.Link), zap.String("query", "/_cluster/initial"), zap.Error(err))
	}
}
func (client *client) queryList() ([]*LiteMachine, error) {
	client.context.GetLog().Debug("查询", zap.String("client.remoteMachine.Link", client.remoteMachine.Link), zap.String("query", "/_cluster/queryMachineList"))
	marshal, err := json.Marshal(client.localMachine.getLiteMachine())
	if err != nil {
		return nil, err
	}
	path := client.remoteMachine.Link + "/_cluster/queryMachineList"
	call, err := client.request.Call(path, marshal)
	if err == nil {
		var liteMachines []*LiteMachine
		err = json.Unmarshal(call, &liteMachines)
		if err == nil {
			return liteMachines, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (client *client) query(parameter *core.Parameter, localValue any) (any, error) {
	path := client.remoteMachine.Link + "/_cluster/query"
	marshal, err := json.Marshal(parameter)
	if err == nil {
		call, err := client.request.Call(path, marshal)
		if err == nil {
			m := util.NewPtr(localValue)
			if len(call) == 0 {
				return m, nil
			}
			err = json.Unmarshal(call, m)
			if err == nil {
				return m, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return nil, err
}

func (client *client) sendAddUser(usernames ...string) {
	path := client.remoteMachine.Link + "/_cluster/addUser"
	us := make([]*User, len(usernames))
	for index, username := range usernames {
		u := NewUser(client.localMachine.MachineId, username)
		us[index] = u
	}
	marshal, err := json.Marshal(us)
	if err == nil {
		client.request.JustCall(path, marshal)
	}
}

func (client *client) sendDeleteUser(usernames ...string) {
	path := client.remoteMachine.Link + "/_cluster/deleteUser"
	us := make([]*User, len(usernames))
	for index, username := range usernames {
		u := NewUser(client.localMachine.MachineId, username)
		us[index] = u
	}
	marshal, err := json.Marshal(us)
	if err == nil {
		client.request.JustCall(path, marshal)
	}
}
func (client *client) sendTextMsg(msg *message.TextMessage) error {
	path := client.remoteMachine.Link + "/_cluster/sendTextMsg"
	marshal, err := json.Marshal(msg)
	if err == nil {
		call, err := client.request.Call(path, marshal)
		if err != nil {
			return err
		} else {
			var response Response
			err := json.Unmarshal(call, &response)
			if err != nil {
				return err
			} else {
				if response.Code == 200 {
					return nil
				} else {
					return core.NoFoundUser
				}
			}

		}
	} else {
		return err
	}
}

func (client *client) initial() error {
	path := client.remoteMachine.Link + "/_cluster/initial"
	marshal, err := json.Marshal(client.localMachine.getLiteMachine())
	if err != nil {
		return err
	} else {
		call, err := client.request.Call(path, marshal)
		if err != nil {
			client.context.GetLog().Error("网络请求错误", zap.Error(err))
			return err
		} else {
			var liteMachine LiteMachine
			err = json.Unmarshal(call, &liteMachine)
			if err != nil {
				client.context.GetLog().Error("网络请求错误", zap.Error(err))
				return err
			}
			client.remoteMachine.MachineId = liteMachine.MachineId
			client.remoteLink = client.remoteMachine.Link
			if liteMachine.MachineId == client.localMachine.MachineId {
				client.isLocal = true
			} else {
				client.isLocal = false
			}
			client.isHandshake = true
			return nil
		}
	}
	return err
}

// HasConn 是否成功建立 链接
func (client *client) HasConn() bool {
	return !client.isLocal && client.isHandshake
}

type store struct {
	tempClientMap *sync.Map
	clientMap     *sync.Map
	num           int
	lock          *sync.Mutex
	localMachine  *Machine
	context       *core.Context
}

func newStore(localMachine *Machine, context *core.Context) *store {
	return &store{clientMap: new(sync.Map), lock: new(sync.Mutex), tempClientMap: new(sync.Map), localMachine: localMachine, context: context}
}

func (s *store) addConfigMachine(remoteMachine *Machine) {
	s.lock.Lock()
	defer s.lock.Unlock()
	client := NewClient(remoteMachine, s.localMachine, s.context)
	client.context.GetLog().Info("添加临时集群节点", zap.String("link", client.remoteLink))
	s.tempClientMap.Store(client.remoteLink, client)
}

func (s *store) addTemp(client *client) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.tempClientMap.Store(client.remoteLink, client)
}
func (s *store) deleteTemp(remoteLink string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.tempClientMap.Delete(remoteLink)
}

// 将机器从临时缓存移动到持久储存
func (s *store) moveTempToStore(client *client) {
	s.lock.Lock()
	defer s.lock.Unlock()
	client.context.GetLog().Info("moveTempToStore", zap.String("link", client.remoteLink), zap.String("MachineId", client.remoteMachine.MachineId))
	s.tempClientMap.Delete(client.remoteLink)
	_, fa := s.clientMap.LoadOrStore(client.remoteMachine.MachineId, client)
	if fa {
		s.num++
	}
}

// 将机器移动至临时缓存
func (s *store) moveStoreToTemp(client *client) {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, fa := s.clientMap.LoadAndDelete(client.remoteMachine.MachineId)
	if fa {
		client.context.GetLog().Info("moveStoreToTemp", zap.String("link", client.remoteLink))
		s.tempClientMap.Store(client.remoteLink, client)
		s.num--
	}
}
func (s *store) getClient(machineId string) (*client, bool) {
	v, ok := s.clientMap.Load(machineId)
	if ok {
		return v.(*client), true
	} else {
		return nil, false
	}
}

func (s *store) eachTempClient(f func(remoteLink string, client *client) bool) {
	s.tempClientMap.Range(func(key, value any) bool {
		return f(key.(string), value.(*client))
	})
}
func (s *store) eachStoreClient(f func(machineId string, client *client) bool) {
	s.clientMap.Range(func(key, value any) bool {
		return f(key.(string), value.(*client))
	})
}

// 如果没有存储机器，则添加为临时机器
func (s *store) addNewClient(client *client) {
	client.context.GetLog().Debug("addNewClient", zap.String("remoteMachine.MachineId", client.remoteMachine.MachineId), zap.String("localMachine.MachineId", s.localMachine.MachineId))
	if client.remoteMachine.MachineId != s.localMachine.MachineId {
		client.context.GetLog().Debug("不是本机添加", zap.String("remoteMachine.MachineId", client.remoteMachine.MachineId), zap.String("localMachine.MachineId", s.localMachine.MachineId))
		s.lock.Lock()
		defer s.lock.Unlock()
		machineId := client.remoteMachine.MachineId
		_, ok := s.clientMap.Load(machineId)
		if !ok {
			client.context.GetLog().Info("addNewClient", zap.String("link", client.remoteLink))
			s.tempClientMap.Store(client.remoteLink, client)
		} else {
			client.context.GetLog().Debug("addNewClient 已存在不添加", zap.String("link", client.remoteLink))
		}
	}
}

type ClientOperate struct {
	localMachine *Machine
	context      *core.Context
	userQueue    *util.Queue
	store        *store
}

func NewClientOperate(context *core.Context, localMachine *Machine) *ClientOperate {
	return &ClientOperate{context: context, localMachine: localMachine, userQueue: util.NewQueue(), store: newStore(localMachine, context)}
}

func (ms *ClientOperate) getClient(machineId string) (*client, bool) {
	return ms.store.getClient(machineId)
}
func (ms *ClientOperate) num() int {
	return ms.store.num
}

// 只用于读取配置文件的时候使用
func (ms *ClientOperate) addConfigMachine(remoteMachine *Machine) {
	ms.store.addConfigMachine(remoteMachine)
}

func (ms *ClientOperate) addNewMachine(machine *Machine) {
	client := NewClient(machine, ms.localMachine, ms.context)
	ms.store.addNewClient(client)

}
func (ms *ClientOperate) sendTextMsg(msg *message.TextMessage, exMachineIds ...string) (string, error) {
	var _err_ = core.NoFoundUser
	machineId := ""
	ms.store.eachStoreClient(func(machineId string, client *client) bool {
		if client.HasConn() && (len(exMachineIds) == 0 || !util.ContainsInArray(exMachineIds, client.remoteMachine.MachineId)) {
			err := client.sendTextMsg(msg)
			_err_ = err
			if err == nil {
				machineId = client.remoteMachine.MachineId
				return false
			}
		}
		return true
	})
	return machineId, _err_
}
func (ms *ClientOperate) Query(parameter *core.Parameter, localValue any) []any {
	vs := make([]any, 0)
	index := 0
	ms.store.eachStoreClient(func(machineId string, client *client) bool {
		if client.HasConn() {
			index++
			parameter.SetString("index", strconv.Itoa(index))
			v, err := client.query(parameter, localValue)
			if err == nil && v != nil {
				v1, ok := v.(*interface{})
				if ok {
					vs = append(vs, *v1)
				} else {
					vs = append(vs, v)
				}

			}
		}
		return true
	})
	return vs
}

func (ms *ClientOperate) initial() {
	ms.store.eachTempClient(func(remoteLink string, client *client) bool {
		err := client.initial()
		if err == nil {
			if client.isLocal {
				ms.context.GetLog().Info("连接到自己，不在尝试连接")
				ms.store.deleteTemp(remoteLink)
			} else if client.HasConn() {
				ms.store.moveTempToStore(client)
			}
		}

		return true
	})
}

func (ms *ClientOperate) live() {
	for {
		time.Sleep(time.Second * 5)
		ms.initial()
		time.Sleep(time.Second * 5)
		ms.store.eachStoreClient(func(machineId string, client *client) bool {
			if client.HasConn() {
				list, err := client.queryList()
				if err != nil {
					if strings.Contains(err.Error(), "Client.Timeout") {
						ms.store.moveStoreToTemp(client)
					}
					ms.context.GetLog().Error("网络请求失败", zap.Error(err))
				} else {
					for _, machine := range list {
						m, err := parseLink(machine.Link)
						if err != nil {
							ms.context.GetLog().Error("parseLink", zap.String("client.remoteLink", client.remoteLink), zap.Error(err))
						} else {
							m.MachineId = machine.MachineId
							client := NewClient(m, ms.localMachine, ms.context)
							ms.store.addNewClient(client)
						}
					}
				}
			}
			return true
		})
	}
}
func (ms *ClientOperate) sendAddUser0(usernames ...string) {
	ms.store.eachStoreClient(func(machineId string, client *client) bool {
		if client.HasConn() {
			client.sendAddUser(usernames...)
		}
		return true
	})
}

func (ms *ClientOperate) sendDeleteUser0(usernames ...string) {
	ms.store.eachStoreClient(func(machineId string, client *client) bool {
		if client.HasConn() {
			client.sendDeleteUser(usernames...)
		}
		return true
	})
}
func (ms *ClientOperate) getMachineLite() []*LiteMachine {
	lm := make([]*LiteMachine, 0)
	ms.store.eachStoreClient(func(machineId string, c *client) bool {
		if c.HasConn() {
			lm = append(lm, c.remoteMachine.getLiteMachine())
		}
		return true
	})
	return lm
}

func (ms *ClientOperate) SendAddUser(username string) {
	ms.userQueue.Offer(&operate{isAdd: true, username: username})
}
func (ms *ClientOperate) SendDeleteUser(username string) {
	ms.userQueue.Offer(&operate{isAdd: false, username: username})
}

func (ms *ClientOperate) userOperate() {
	deleteUsers := make([]string, 0)
	addUsers := make([]string, 0)
	for {
		v, num := ms.userQueue.Poll()
		op := v.(*operate)
		if op.isAdd {
			addUsers = append(addUsers, op.username)
		} else {
			deleteUsers = append(deleteUsers, op.username)
		}
		if num == 0 || num >= 100 {
			if len(addUsers) > 0 {
				ms.sendAddUser0(addUsers...)
				addUsers = make([]string, 0)
			}
			if len(deleteUsers) > 0 {
				ms.sendDeleteUser0(deleteUsers...)
				deleteUsers = make([]string, 0)
			}
		}
	}
}
func (ms *ClientOperate) run() {
	ms.initial()
	go ms.live()
	go ms.userOperate()
}
