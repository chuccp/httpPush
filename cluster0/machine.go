package cluster0

import (
	"net"
	"net/http"
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

type MultiMessage struct {
	From   string
	ToUser []string
	Text   string
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
func parseLiteMachine(liteMachine *LiteMachine, re *http.Request) (*Machine, error) {
	url, err := url.Parse(liteMachine.Link)
	if err != nil {
		return nil, err
	}
	var machine Machine
	machine.MachineId = liteMachine.MachineId
	machine.Scheme = url.Scheme
	machine.Address = url.Hostname()
	ip := net.ParseIP(machine.Address)
	if ip == nil || ip.Equal(net.IPv4zero) || ip.Equal(net.IPv6zero) {
		addr, _, err := net.SplitHostPort(re.RemoteAddr)
		if err != nil {
			return nil, err
		}
		machine.Address = addr
	}
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return nil, err
	}
	machine.Port = port
	machine.Link = url.Scheme + "://" + machine.Address + ":" + url.Port()
	return &machine, nil
}
