package cluster

import (
	"encoding/json"
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

func (client *client) initial() error {
	log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	path := client.remoteMachine.Link + "/_cluster/initial"
	marshal, err := json.Marshal(client.localMachine.getLiteMachine())
	if err != nil {
		return err
	} else {
		log.Println("@@@@@@@@@@@@@@@@@@@!!!!!!!!", path, string(marshal))
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
}

func NewClientStore(localMachine *Machine) *ClientStore {
	return &ClientStore{clientMap: new(sync.Map), lock: new(sync.Mutex), tempClientMap: new(sync.Map), localMachine: localMachine}
}

// 只用于读取配置文件的时候使用
func (ms *ClientStore) addMachineNoMachineId(remoteMachine *Machine) {
	ms.lock.Lock()
	defer ms.lock.Unlock()
	client := NewClient(remoteMachine, ms.localMachine)
	ms.tempClientMap.Store(client.remoteLink, client)
}
func (ms *ClientStore) addNewMachine(machineId string, machine *Machine) {

	log.Println("addNewMachine", machineId)
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
		} else {
			log.Println("initial===========", client.remoteLink, err)
		}
		return true
	})
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
	//time.Sleep(2 * time.Second)
	ms.initial()
	go ms.live()
}
