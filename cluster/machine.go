package cluster

import (
	"log"
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
func parseLink2(link string, re *http.Request) (*Machine, error) {
	url, err := url.Parse(link)
	if err != nil {
		return nil, err
	}
	var machine Machine
	machine.Scheme = url.Scheme
	machine.Address = url.Hostname()
	if machine.Address == net.IPv4zero.String() {
		addr, _, err := net.SplitHostPort(re.RemoteAddr)
		if err != nil {
			return nil, err
		}
		machine.Address = addr
	}
	log.Println("parseLink2:", machine.Address)
	port, err := strconv.Atoi(url.Port())
	if err != nil {
		return nil, err
	}
	machine.Port = port
	machine.Link = url.Scheme + "//" + machine.Address + ":" + url.Port()
	log.Println("machine.Link:" + machine.Link)
	return &machine, nil
}
