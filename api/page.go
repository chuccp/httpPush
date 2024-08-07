package api

import (
	"github.com/chuccp/httpPush/util"
	"strings"
	"time"
)

type Page struct {
	Num  int
	List []*PageUser
}

func NewPage() *Page {
	return &Page{0, make([]*PageUser, 0)}
}
func (p *Page) AddPageUser(list []*PageUser) {
	p.List = append(p.List, list...)
}
func (p *Page) AddPage(page *Page) {
	p.List = append(p.List, page.List...)
	p.Num = p.Num + page.Num
}
func (p *Page) AddNum(num int) {
	p.Num = p.Num + num
}

type PageUser struct {
	UserName       string
	MachineAddress string
	CreateTime     string
	MachineId      string
	Conn           []*Conn
}
type ResponseMsg struct {
	Time   string
	Result string
}

func NewResponseMsg(fa bool) *ResponseMsg {
	t := time.Now()
	if fa {
		return &ResponseMsg{Time: util.FormatTime(&t), Result: "success"}
	}
	return &ResponseMsg{Time: util.FormatTime(&t), Result: "offline"}
}

type GroupMsg struct {
	MachineAddress string
	Num            int32
	MachineId      string
}
type OrderUser struct {
	Priority  int
	MachineId string
	OrderTime string
}

type ClusterUserNum struct {
	UserNum   any
	MachineId string
}

type AllOrderUser struct {
	OrderUser      []*OrderUser
	MachineId      string
	MachineAddress string
}
type TimeWheelLogsByAsc []*TimeWheelLog

func (p TimeWheelLogsByAsc) Len() int {
	return len(p)
}
func (p TimeWheelLogsByAsc) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
func (p TimeWheelLogsByAsc) Less(i, j int) bool {
	return strings.Compare(p[i].StartTime, p[j].StartTime) < 0
}

type PageTimeWheelLog struct {
	MachineAddress string
	MachineId      string
	TimeWheelLogs  []*TimeWheelLog
}

type GroupInfo struct {
	MachineAddress string
	MachineId      string
	GroupInfo      map[string]int
}

type TimeWheelLog struct {
	Num       int
	Cha       int
	StartTime string
	EndTime   string
}

func NewAllOrderUser() *AllOrderUser {
	return &AllOrderUser{OrderUser: make([]*OrderUser, 0)}
}

type Version struct {
	Version        string
	StartTime      string
	MachineId      string
	MachineAddress string
}
