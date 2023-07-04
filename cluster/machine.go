package cluster

import (
	"net/url"
	"strconv"
	"sync"
)

type Machine struct {
	Link      string
	MachineId string
	IsLocal   bool
	Scheme    string
	Address   string
	Port      int
}

type MachineStore struct {
	linkMachineMap      *sync.Map
	machineIdMachineMap *sync.Map
}

func (ms *MachineStore) addMachineByLink(machine *Machine) {
	ms.linkMachineMap.Store(machine.Link, machine)
}
func (ms *MachineStore) addMachineById(machine *Machine) {
	ms.machineIdMachineMap.Store(machine.MachineId, machine)

}
func (ms *MachineStore) RangeMachineByLink(f func(link string, machine *Machine) bool) {
	ms.linkMachineMap.Range(func(key, value any) bool {
		return f(key.(string), value.(*Machine))
	})
}

func NewMachineStore() *MachineStore {
	return &MachineStore{linkMachineMap: new(sync.Map), machineIdMachineMap: new(sync.Map)}
}
func parseLink(link string) (*Machine, error) {
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	var machine Machine
	machine.Scheme = url.Scheme
	machine.Address = url.Hostname()
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return nil, err
	}
	machine.Port = port
	machine.Link = link
	return &machine, nil
}
