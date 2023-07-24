package cluster

import (
	"net/url"
	"strconv"
)

type Machine struct {
	Link        string
	MachineId   string
	IsLocal     bool
	Scheme      string
	Address     string
	Port        int
	liteMachine *LiteMachine
}
type LiteMachine struct {
	Link      string
	MachineId string
}

func (machine *Machine) getLiteMachine() *LiteMachine {
	if machine.liteMachine != nil {
		return machine.liteMachine
	} else {
		machine.liteMachine = &LiteMachine{Link: machine.Link, MachineId: machine.MachineId}
	}
	return machine.liteMachine
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
