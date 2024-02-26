package api

import (
	"github.com/chuccp/httpPush/util"
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
type AllOrderUser struct {
	OrderUser      []*OrderUser
	MachineId      string
	MachineAddress string
}

func NewAllOrderUser() *AllOrderUser {
	return &AllOrderUser{OrderUser: make([]*OrderUser, 0)}
}
