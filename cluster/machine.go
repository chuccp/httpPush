package cluster

import (
	"net/url"
	"strconv"
	"sync"
)

type Machine struct {
	MachineId string
	IsLocal   bool
	Scheme    string
	Address   string
	Port      int
}

type MachineStore struct {
	machineMap *sync.Map
}

func NewMachineStore() *MachineStore {
	return &MachineStore{machineMap: new(sync.Map)}
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
	return &machine, nil
}
