package cluster

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/util"
	"log"
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

func NewClient(remoteMachine *Machine, localMachine *Machine) *client {
	return &client{request: util.NewRequest(), isHandshake: false, remoteMachine: remoteMachine, localMachine: localMachine}
}
func (client *client) run() {
	log.Println("握手", client.remoteMachine.Link, "/_cluster/initial")
	err := client.initial()
	if err == nil {
		log.Println("握手", client.remoteMachine.Link, "/_cluster/initial", "完成")
	} else {
		log.Println("握手", client.remoteMachine.Link, "/_cluster/initial", "失败", err)
	}
}
func (client *client) queryList() ([]*LiteMachine, error) {
	log.Println("心跳 交换数据", client.remoteMachine.Link, "/_cluster/queryMachineList")
	path := client.remoteMachine.Link + "/_cluster/queryMachineList"
	call, err := client.request.Get(path)
	if err == nil {
		log.Println("心跳 交换数据", string(call))
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
	log.Println("查询接口", client.remoteMachine.Link, "/_cluster/query")
	path := client.remoteMachine.Link + "/_cluster/query"
	marshal, err := json.Marshal(parameter)
	if err == nil {
		call, err := client.request.Call(path, marshal)
		if err == nil {
			m := util.NewPtr(localValue)
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
func (client *client) sendDeleteUser(username string) {
	path := client.remoteMachine.Link + "/_cluster/deleteUser"
	u := NewUser(client.localMachine.MachineId, username)
	marshal, err := json.Marshal(u)
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
				if response.code == 200 {
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
		log.Println("!!!!!!!!!!!!", err)
		if err != nil {
			return err
		} else {
			var liteMachine LiteMachine
			err = json.Unmarshal(call, &liteMachine)
			if err != nil {
				return err
			}
			client.remoteMachine.MachineId = liteMachine.MachineId
			client.remoteLink = liteMachine.Link
			if liteMachine.MachineId == client.localMachine.MachineId {
				client.isLocal = true
			} else {
				client.isLocal = false
			}

			return nil
		}
	}
	return err
}

type ClientStore struct {
	tempClientMap *sync.Map
	clientMap     *sync.Map
	lock          *sync.Mutex
	localMachine  *Machine
	context       *core.Context
}

func NewClientStore(context *core.Context, localMachine *Machine) *ClientStore {
	return &ClientStore{context: context, clientMap: new(sync.Map), lock: new(sync.Mutex), tempClientMap: new(sync.Map), localMachine: localMachine}
}

func (ms *ClientStore) getClient(machineId string) (*client, bool) {
	v, ok := ms.clientMap.Load(machineId)
	if ok {
		return v.(*client), true
	} else {
		return nil, false
	}
}

// 只用于读取配置文件的时候使用
func (ms *ClientStore) addMachineNoMachineId(remoteMachine *Machine) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	client := NewClient(remoteMachine, ms.localMachine)
	ms.tempClientMap.Store(client.remoteLink, client)
}
func (ms *ClientStore) addNewMachine(machineId string, machine *Machine) {
	if machineId != ms.localMachine.MachineId {
		_, ok := ms.clientMap.Load(machineId)
		if !ok {
			client := NewClient(machine, ms.localMachine)
			err := client.initial()
			if err != nil {
				log.Println("initial", machineId, client.remoteLink, err)
			} else {
				ms.addMachineClient(machineId, client)
			}
		}
	} else {
		log.Println("自己连接到自己，不处理")
	}
}
func (ms *ClientStore) addMachineClient(machineId string, client *client) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	_, ok := ms.clientMap.Load(machineId)
	if !ok {
		ms.clientMap.LoadOrStore(machineId, client)
	}
}
func (ms *ClientStore) getMachineLite() []*LiteMachine {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	lm := make([]*LiteMachine, 0)
	ms.clientMap.Range(func(key, value any) bool {
		c := value.(*client)
		lm = append(lm, c.remoteMachine.getLiteMachine())
		return true
	})
	return lm
}
func (ms *ClientStore) initial() {
	ms.tempClientMap.Range(func(k, value any) bool {
		client := value.(*client)
		err := client.initial()
		if err == nil {
			if !client.isLocal {
				ms.tempClientMap.Delete(k)
				ms.addMachineClient(client.remoteMachine.MachineId, client)
			}
		}
		return true
	})
}

func (ms *ClientStore) sendTextMsg(msg *message.TextMessage, exMachineId string) (string, error) {
	var _err_ error = core.NoFoundUser
	machineId := ""
	ms.clientMap.Range(func(k, value any) bool {
		client := value.(*client)
		if !client.isLocal && client.remoteMachine.MachineId != exMachineId {
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

func (ms *ClientStore) SendAddUser(username string) {
	ms.clientMap.Range(func(k, value any) bool {
		client := value.(*client)
		if !client.isLocal {
			client.sendAddUser(username)
		}
		return true
	})
}
func (ms *ClientStore) SendDeleteUser(username string) {
	ms.clientMap.Range(func(k, value any) bool {
		client := value.(*client)
		if !client.isLocal {
			client.sendDeleteUser(username)
		}
		return true
	})
}

func (ms *ClientStore) Query(parameter *core.Parameter, localValue any) []any {
	vs := make([]any, 0)
	ms.clientMap.Range(func(k, value any) bool {
		client := value.(*client)
		if !client.isLocal {
			v, err := client.query(parameter, localValue)
			if err == nil && v != nil {
				vs = append(vs, v)
			}
		}
		return true
	})
	return vs
}

func (ms *ClientStore) live() {
	for {
		time.Sleep(time.Minute)
		ms.clientMap.Range(func(_, value any) bool {
			client := value.(*client)
			list, err := client.queryList()
			if err != nil {
				log.Println("queryList", client.remoteLink, err)
			} else {
				for _, machine := range list {
					m, err := parseLink(machine.Link)
					if err != nil {
						log.Println("parseLink", client.remoteLink, err)
					} else {
						client := NewClient(m, ms.localMachine)
						ms.addMachineClient(machine.MachineId, client)
					}
				}
			}
			return true
		})
	}
}
func (ms *ClientStore) run() {
	ms.initial()
	go ms.live()
}
