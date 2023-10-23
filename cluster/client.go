package cluster

import (
	"encoding/json"
	"fmt"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/util"
	"log"
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
}

type operate struct {
	isAdd    bool
	username string
}

func NewClient(remoteMachine *Machine, localMachine *Machine) *client {
	return &client{request: util.NewRequest(), isHandshake: false, remoteMachine: remoteMachine, localMachine: localMachine}
}
func (client *client) run() {
	err := client.initial()
	if err == nil {
		log.Println("握手", client.remoteMachine.Link, "/_cluster/initial", "完成")
	} else {
		log.Println("握手", client.remoteMachine.Link, "/_cluster/initial", "失败", err)
	}
}
func (client *client) queryList() ([]*LiteMachine, error) {
	log.Println("查询", client.remoteMachine.Link, "/_cluster/queryMachineList")
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
			return fmt.Errorf("initial Call err:%s", err.Error())
		} else {
			var liteMachine LiteMachine
			err = json.Unmarshal(call, &liteMachine)
			if err != nil {
				return fmt.Errorf("initial Call json.Unmarshal err:%s", err.Error())
			}
			client.remoteMachine.MachineId = liteMachine.MachineId
			client.remoteLink = liteMachine.Link
			if liteMachine.MachineId == client.localMachine.MachineId {
				client.isLocal = true
			} else {
				client.isLocal = false
			}
			log.Println("path:", path, " body:", string(marshal), " back:", string(call), "  isLocal:", client.isLocal)
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
}

func newStore(localMachine *Machine) *store {
	return &store{clientMap: new(sync.Map), lock: new(sync.Mutex), tempClientMap: new(sync.Map), localMachine: localMachine}
}

func (s *store) addConfigMachine(remoteMachine *Machine) {
	s.lock.Lock()
	defer s.lock.Unlock()
	client := NewClient(remoteMachine, s.localMachine)
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
	if client.remoteMachine.MachineId != s.localMachine.MachineId {
		s.lock.Lock()
		defer s.lock.Unlock()
		machineId := client.remoteMachine.MachineId
		_, ok := s.clientMap.Load(machineId)
		if !ok {
			s.tempClientMap.Store(client.remoteLink, client)
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
	return &ClientOperate{context: context, localMachine: localMachine, userQueue: util.NewQueue(), store: newStore(localMachine)}
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
	client := NewClient(machine, ms.localMachine)
	ms.store.addNewClient(client)

}
func (ms *ClientOperate) sendTextMsg(msg *message.TextMessage, exMachineId string) (string, error) {
	var _err_ error = core.NoFoundUser
	machineId := ""
	ms.store.eachStoreClient(func(machineId string, client *client) bool {
		if client.HasConn() && client.remoteMachine.MachineId != exMachineId {
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
				vs = append(vs, v)
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
				log.Println("连接到自己，不在尝试连接")
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
					err = fmt.Errorf("queryList:%s err:%s", client.remoteLink, err)
					log.Println(err)
				} else {
					for _, machine := range list {
						m, err := parseLink(machine.Link)
						if err != nil {
							log.Println("parseLink", client.remoteLink, err)
						} else {
							client := NewClient(m, ms.localMachine)
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
