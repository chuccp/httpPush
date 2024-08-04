package cluster

import (
	"encoding/json"
	"errors"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Machine struct {
	Link      string
	MachineId string
}

type Machines struct {
	machines []*Machine
}

func NewMachines() *Machines {
	return &Machines{machines: make([]*Machine, 0)}
}
func (m *Machines) addFirstMachine(machine *Machine) {
	for _, m2 := range m.machines {
		if m2.Link == machine.Link {
			return
		}
	}
	m.machines = append(m.machines, machine)
}
func (m *Machines) addMachine(machine *Machine) {
	for _, m2 := range m.machines {
		if machine.MachineId == m2.MachineId || m2.Link == machine.Link {
			return
		}
	}
	m.machines = append(m.machines, machine)
}
func (m *Machines) removeMachine(machine *Machine) {
	data := make([]*Machine, 0)
	for _, m2 := range m.machines {
		if machine.MachineId == m2.MachineId || m2.Link == machine.Link {
			return
		} else {
			data = append(data, m2)
		}
	}
	m.machines = data
}

func (m *Machines) hasMachine(machine *Machine) bool {
	for _, m2 := range m.machines {
		if machine.MachineId == m2.MachineId || m2.Link == machine.Link {
			return true
		}
	}
	return false
}
func (m *Machines) getMachineLink(machineId string) string {
	for _, machine := range m.machines {
		if machine.MachineId == machineId {
			return machine.Link
		}
	}
	return ""
}
func (m *Machines) getMachines() []*Machine {
	data := m.machines
	if len(data) == 0 {
		return make([]*Machine, 0)
	}
	machines := make([]*Machine, len(data))
	copy(machines, data)
	return machines
}
func (m *Machines) getExMachines(machineIds ...string) []*Machine {
	data := m.machines
	if len(data) == 0 {
		return make([]*Machine, 0)
	}
	machines := make([]*Machine, 0)
	for _, machine := range m.machines {
		if !util.ContainsArray(machine.MachineId, machineIds...) {
			machines = append(machines, machine)
		}
	}
	return machines
}

type MachineStore struct {
	tempMachines *Machines
	machines     *Machines
	lock         *sync.RWMutex
	httpClient   *HttpClient
	localMachine *Machine
	context      *core.Context
}

func NewMachineStore(context *core.Context) *MachineStore {
	return &MachineStore{tempMachines: NewMachines(), machines: NewMachines(), lock: new(sync.RWMutex), httpClient: NewHttpClient(), context: context}
}

func (machineStore *MachineStore) addFirstMachine(machine *Machine) {
	machineStore.lock.Lock()
	defer machineStore.lock.Unlock()
	machineStore.tempMachines.addFirstMachine(machine)
}

func (machineStore *MachineStore) addMachine(machine *Machine) {
	machineStore.lock.Lock()
	defer machineStore.lock.Unlock()
	if len(machine.MachineId) == 0 || machine.MachineId == machineStore.localMachine.MachineId {
		return
	}
	if !machineStore.machines.hasMachine(machine) {
		machineStore.tempMachines.addMachine(machine)
	}
}
func (machineStore *MachineStore) addMachines(machines []*Machine) {
	machineStore.lock.Lock()
	defer machineStore.lock.Unlock()
	for _, machine := range machines {
		if len(machine.MachineId) == 0 {
			return
		}
		if !machineStore.machines.hasMachine(machine) {
			machineStore.tempMachines.addMachine(machine)
		}
	}
}

func (machineStore *MachineStore) GetTempMachines() []*Machine {
	machineStore.lock.RLock()
	defer machineStore.lock.RUnlock()
	return machineStore.tempMachines.getMachines()
}
func (machineStore *MachineStore) GetMachines() []*Machine {
	machineStore.lock.RLock()
	defer machineStore.lock.RUnlock()
	return machineStore.machines.getMachines()
}
func (machineStore *MachineStore) GetMachineNum() int {
	machineStore.lock.RLock()
	defer machineStore.lock.RUnlock()
	return len(machineStore.machines.machines)
}
func (machineStore *MachineStore) removeTempToMachines(machine *Machine) {
	machineStore.lock.Lock()
	defer machineStore.lock.Unlock()
	machineStore.tempMachines.removeMachine(machine)
	machineStore.machines.addMachine(machine)
}
func (machineStore *MachineStore) removeTempMachine(machine *Machine) {
	machineStore.lock.Lock()
	defer machineStore.lock.Unlock()
	machineStore.tempMachines.removeMachine(machine)
}

func (machineStore *MachineStore) getExMachines(machineIds ...string) []*Machine {
	machineStore.lock.RLock()
	defer machineStore.lock.RUnlock()
	return machineStore.machines.getExMachines(machineIds...)
}

func (machineStore *MachineStore) Query(parameter *core.Parameter, localValue any) []any {
	vs := make([]any, 0)
	index := 0
	waitGroup := new(sync.WaitGroup)
	var lock sync.Mutex
	machines := machineStore.GetMachines()

	for _, machine := range machines {
		parameter.SetIndex(index)
		data, err := json.Marshal(parameter)
		if err == nil {
			waitGroup.Add(1)
			go func(jsonData []byte) {
				defer waitGroup.Done()
				v, err := machineStore.query(machine, jsonData, localValue)
				if err != nil {
					machineStore.context.GetLog().Error("query", zap.Error(err), zap.Any("value", v))
				}
				if err == nil && v != nil {
					v1, ok := v.(*any)
					if ok {
						lock.Lock()
						vs = append(vs, *v1)
						lock.Unlock()
					} else {
						lock.Lock()
						vs = append(vs, v)
						lock.Unlock()
					}
				}

			}(data)
		}

	}

	waitGroup.Wait()
	return vs
}

func (machineStore *MachineStore) GetMachineLink(machineId string) string {
	machineStore.lock.RLock()
	defer machineStore.lock.RUnlock()
	return machineStore.machines.getMachineLink(machineId)
}
func (machineStore *MachineStore) initial(machine *Machine, data []byte) {
	machineStore.context.GetLog().Debug("initial", zap.Any("machine", machine))
	call, err := machineStore.httpClient.Call(machine, "/_cluster/initial", data)
	if err == nil {
		var _machine_ Machine
		err = json.Unmarshal(call, &_machine_)
		if _machine_.MachineId != machineStore.localMachine.MachineId {
			machine.MachineId = _machine_.MachineId
			machineStore.removeTempToMachines(machine)
		} else {
			machineStore.context.GetLog().Info("initial 握手的对象是自己移除", zap.String("MachineId", _machine_.MachineId))
			machine.MachineId = _machine_.MachineId
			machineStore.removeTempMachine(machine)
		}
	} else {
		machineStore.context.GetLog().Error("initial", zap.Any("machine", machine), zap.Error(err))
	}
}

func (machineStore *MachineStore) query(machine *Machine, data []byte, localValue any) (any, error) {
	call, err := machineStore.httpClient.Call(machine, "/_cluster/query", data)
	if err == nil {
		if len(call) == 0 {
			return nil, errors.New("NO_VALUE")
		}
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
func (machineStore *MachineStore) queryMachines(machine *Machine, data []byte) {
	call, err := machineStore.httpClient.Call(machine, "/_cluster/queryMachineList", data)
	if err == nil {
		var machines []*Machine
		err = json.Unmarshal(call, &machines)
		if err == nil {
			machineStore.addMachines(machines)
		}
	} else {
		machineStore.context.GetLog().Error("queryMachines", zap.Any("machine", machine), zap.Error(err))
	}
}
func (machineStore *MachineStore) sendMsg(machineId string, data []byte) error {
	link := machineStore.GetMachineLink(machineId)
	if len(link) > 0 {
		call, err := machineStore.httpClient.CallByLink(link, "/_cluster/sendTextMsg", data)
		if err == nil {
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
		} else {
			return err
		}
	} else {
		return core.NoFoundUser
	}
}

func (machineStore *MachineStore) initials() {
	data, _ := json.Marshal(machineStore.localMachine)
	for _, machine := range machineStore.GetTempMachines() {
		machineStore.initial(machine, data)
		time.Sleep(time.Second)
	}
}
func (machineStore *MachineStore) queryMachineList() {
	data, _ := json.Marshal(machineStore.localMachine)
	for _, machine := range machineStore.GetMachines() {
		machineStore.queryMachines(machine, data)
		time.Sleep(time.Second)
	}
}

func (machineStore *MachineStore) num() int {
	return machineStore.GetMachineNum()
}
