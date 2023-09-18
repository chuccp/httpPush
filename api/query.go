package api

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"log"
	"net/http"
)

type Query struct {
	context *core.Context
	server  core.IHttpServer
}

func (query *Query) Init() {
	query.AddQuery("/queryUser", query.queryUser, query.queryUserApi)
	query.AddQuery("/onlineUser", query.onlineUser, query.onlineUserApi)
	query.AddQuery("/sendGroupMsg", query.sendGroupMsg, query.sendGroupMsgApi)
}
func (query *Query) AddQuery(handleName string, handle core.RegisterHandle, handler func(http.ResponseWriter, *http.Request)) {
	query.context.RegisterHandle(handleName, handle)
	query.server.AddHttpRoute(handleName, handler)
}
func (query *Query) queryUserApi(w http.ResponseWriter, re *http.Request) {
	parameter := core.NewParameter(re)
	value := query.context.Query(parameter)
	data, _ := json.Marshal(value)
	w.Write(data)
}
func (query *Query) queryUser(parameter *core.Parameter) any {
	var u User
	machineInfoId, ok := query.context.GetHandle("machineInfoId")
	if ok {
		u.Machine = machineInfoId(parameter)
	}
	u.Conn = make([]*Conn, 0)
	data, _ := json.Marshal(parameter)
	log.Println(string(data))
	id := core.GetUsername(parameter)
	log.Println("id", id)
	if len(id) > 0 {
		us, fa := query.context.GetUser(id)
		if fa {
			for _, iUser := range us {
				u.Username = iUser.GetUsername()
				u.Conn = append(u.Conn, newConn(iUser.GetRemoteAddress(), iUser.LastLiveTime().Format(util.TimestampFormat), iUser.CreateTime().Format(util.TimestampFormat)))
			}
		}
	}
	return &u
}
func (query *Query) onlineUserApi(writer http.ResponseWriter, request *http.Request) {
	parameter := core.NewParameter(request)
	page := NewPage()
	values := query.context.Query(parameter).([]any)
	for _, value := range values {
		p := value.(*Page)
		page.AddPage(p)
	}
	machineAddress, ok := query.context.GetHandle("machineAddress")
	if ok {
		for _, pageUser := range page.List {
			parameter.SetString("machineId", pageUser.MachineId)
			pageUser.MachineAddress = machineAddress(parameter).(string)
		}
	}
	data, _ := json.Marshal(page)
	writer.Write(data)
}
func (query *Query) onlineUser(parameter *core.Parameter) any {
	handle, ok := query.context.GetHandle("remoteMachineNum")
	total := 1
	if ok {
		num := handle(parameter).(int)
		total = total + num
	}
	size := parameter.GetInt("size")
	if size < 1 {
		size = 10
	}
	index := parameter.GetInt("index")
	num := size / total
	yu := size % total
	if yu > index {
		num = num + 1
	}
	pageUsers := make([]*PageUser, 0)
	machineId := ""
	machineInfoId, ok := query.context.GetHandle("machineInfoId")
	if ok {
		machineId = machineInfoId(parameter).(string)
	}
	if num > 0 {
		query.context.RangeUser(func(username string, user *user.StoreUser) bool {
			num--
			pageUsers = append(pageUsers, &PageUser{UserName: username, MachineId: machineId, CreateTime: user.GetCreateTime()})
			return num > 0
		})
	}
	return &Page{List: pageUsers, Num: query.context.GetUserNum()}
}

func (query *Query) sendGroupMsg(parameter *core.Parameter) any {
	groupId := parameter.GetString("groupId")
	msg := parameter.GetString("msg")
	groupMsg := &GroupMsg{}
	machineInfoId, ok := query.context.GetHandle("machineInfoId")
	if ok {
		groupMsg.MachineId = machineInfoId(parameter).(string)
	}
	groupMsg.Num = query.context.SendGroupTextMessage("system", groupId, msg)
	return groupMsg
}
func (query *Query) sendGroupMsgApi(writer http.ResponseWriter, request *http.Request) {
	parameter := core.NewParameter(request)
	daMap := make(map[string]any)
	var total int32 = 0
	values := query.context.Query(parameter).([]any)
	machineAddress, ok := query.context.GetHandle("machineAddress")
	for _, value := range values {
		p := value.(*GroupMsg)
		total = p.Num + total
		if ok {
			parameter.SetString("machineId", p.MachineId)
			p.MachineAddress = machineAddress(parameter).(string)
		}
	}
	daMap["total"] = total
	daMap["list"] = values
	data, _ := json.Marshal(daMap)
	writer.Write(data)
}

func NewQuery(context *core.Context, server core.IHttpServer) *Query {
	query := &Query{context: context, server: server}
	return query
}

type User struct {
	Username string
	Conn     []*Conn
	Machine  any
}
type Conn struct {
	RemoteAddress string
	LastLiveTime  string
	CreateTime    string
}

func newConn(RemoteAddress string, LastLiveTime string, CreateTime string) *Conn {
	return &Conn{RemoteAddress, LastLiveTime, CreateTime}
}
