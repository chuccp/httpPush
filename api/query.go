package api

import (
	"encoding/json"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"net/http"
	"sort"
)

type Query struct {
	context *core.Context
	server  core.IHttpServer
}

func (query *Query) Init() {
	query.AddQuery("/queryUser", query.queryUser, query.queryUserApi)
	query.AddQuery("/queryHistory", query.queryHistory, query.queryHistoryApi)
	query.AddQuery("/onlineUser", query.onlineUser, query.onlineUserApi)
	query.AddQuery("/sendGroupMsg", query.sendGroupMsg, query.sendGroupMsgApi)
	query.AddQuery("/info_user", query.clusterInfo, query.clusterInfoApi)
	query.AddQuery("/queryOrderInfo", query.queryOrderInfo, query.queryOrderInfoApi)

}

func (query *Query) clusterInfoApi(writer http.ResponseWriter, request *http.Request) {
	machineInfoView := &MachineInfos{Total: 0, List: make([]*MachineInfo, 0)}
	parameter := core.NewParameter(request)
	values := query.context.Query(parameter).([]any)
	for _, value := range values {
		machineInfo := value.(*MachineInfo)
		machineInfo.Address, _ = query.getMachineAddress(machineInfo.MachineId, parameter)
		machineInfoView.List = append(machineInfoView.List, machineInfo)
		machineInfoView.Total = machineInfoView.Total + machineInfo.UserNum
	}
	data, _ := json.Marshal(machineInfoView)
	writer.Write(data)
}

type MachineInfo struct {
	Address   string
	UserNum   int
	MachineId string
}
type MachineInfos struct {
	List  []*MachineInfo `json:"cluster"`
	Total int            `json:"total"`
}

func (query *Query) clusterInfo(parameter *core.Parameter) any {
	var machineInfo MachineInfo
	machineInfo.UserNum = query.context.GetUserNum()
	machineId, _ := query.getMachineInfoId(parameter)
	machineInfo.MachineId = machineId
	return &machineInfo
}

func (query *Query) AddQuery(handleName string, handle core.RegisterHandle, handler func(http.ResponseWriter, *http.Request)) {
	query.context.RegisterHandle(handleName, handle)
	query.server.AddHttpRoute(handleName, handler)
}

func (query *Query) queryHistory(parameter *core.Parameter) any {
	id := core.GetUsername(parameter)
	var history = &History{}
	log, fa := query.context.GetHistory(id)
	if fa {
		history.History = log
	}
	machineInfoId, ok := query.getMachineInfoId(parameter)
	if ok {
		history.Machine = machineInfoId
	}
	return history
}

func (query *Query) queryHistoryApi(writer http.ResponseWriter, request *http.Request) {
	parameter := core.NewParameter(request)
	value := query.context.Query(parameter)
	data, _ := json.Marshal(value)
	writer.Write(data)
}

func (query *Query) queryUserApi(w http.ResponseWriter, re *http.Request) {
	parameter := core.NewParameter(re)
	value := query.context.Query(parameter)
	data, _ := json.Marshal(value)
	w.Write(data)
}
func (query *Query) queryUser(parameter *core.Parameter) any {
	var u User
	machineInfoId, ok := query.getMachineInfoId(parameter)
	if ok {
		u.Machine = machineInfoId
	}
	u.Conn = make([]*Conn, 0)
	id := core.GetUsername(parameter)
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
	for _, pageUser := range page.List {
		machineAddress, ok := query.getMachineAddress(pageUser.MachineId, parameter)
		if ok {
			pageUser.MachineAddress = machineAddress
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
	num := 0
	yu := size
	if total != 0 {
		num = size / total
		yu = size % total
	}
	if yu > index {
		num = num + 1
	}
	pageUsers := make([]*PageUser, 0)
	machineInfoId, _ := query.getMachineInfoId(parameter)
	if num > 0 {
		query.context.RangeUser(func(username string, user *user.StoreUser) bool {
			num--
			pageUsers = append(pageUsers, &PageUser{UserName: username, MachineId: machineInfoId, CreateTime: user.GetCreateTime()})
			return num > 0
		})
	}
	return &Page{List: pageUsers, Num: query.context.GetUserNum()}
}
func (query *Query) getMachineInfoId(parameter *core.Parameter) (string, bool) {
	machineInfoId, ok := query.context.GetHandle("machineInfoId")
	if ok {
		mm := machineInfoId(parameter)
		return mm.(string), ok
	}
	return "", false
}

func (query *Query) getMachineAddress(machineInfoId string, parameter *core.Parameter) (string, bool) {
	machineAddress, ok := query.context.GetHandle("machineAddress")
	if ok {
		parameter.SetString("machineId", machineInfoId)
		return machineAddress(parameter).(string), true
	}
	return "", false
}

func (query *Query) sendGroupMsg(parameter *core.Parameter) any {
	groupId := parameter.GetString("groupId")
	msg := parameter.GetString("msg")
	groupMsg := &GroupMsg{}
	machineInfoId, ok := query.getMachineInfoId(parameter)
	if ok {
		groupMsg.MachineId = machineInfoId
	}
	groupMsg.Num = query.context.SendGroupTextMessage("system", groupId, msg)
	return groupMsg
}
func (query *Query) sendGroupMsgApi(writer http.ResponseWriter, request *http.Request) {
	parameter := core.NewParameter(request)
	daMap := make(map[string]any)
	var total int32 = 0
	values := query.context.Query(parameter).([]any)
	for _, value := range values {
		p := value.(*GroupMsg)
		total = p.Num + total
		machineAddress, ok := query.getMachineAddress(p.MachineId, parameter)
		if ok {
			p.MachineAddress = machineAddress
		}
	}
	daMap["total"] = total
	daMap["list"] = values
	data, _ := json.Marshal(daMap)
	writer.Write(data)
}

func (query *Query) queryOrderInfo(parameter *core.Parameter) any {
	userId := parameter.GetVString("userId", "username", "id")
	us := query.context.GetUserAllOrder(userId)
	sort.Sort(user.ByAsc(us))
	allOrderUser := NewAllOrderUser()
	machineInfoId, ok := query.getMachineInfoId(parameter)
	if ok {
		allOrderUser.MachineId = machineInfoId
	}
	ous := make([]*OrderUser, 0)
	for _, u := range us {
		ous = append(ous, &OrderUser{Priority: u.GetPriority(), MachineId: u.GetMachineId(), OrderTime: util.FormatTime(u.GetOrderTime())})
	}
	allOrderUser.OrderUser = ous
	return allOrderUser
}

func (query *Query) queryOrderInfoApi(writer http.ResponseWriter, request *http.Request) {
	parameter := core.NewParameter(request)
	values := query.context.Query(parameter).([]any)
	for _, value := range values {
		p := value.(*AllOrderUser)
		machineAddress, ok := query.getMachineAddress(p.MachineId, parameter)
		if ok {
			p.MachineAddress = machineAddress
		}
	}
	data, _ := json.Marshal(values)
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

type History struct {
	Machine string
	History *user.HistoryMessage
}

func newConn(RemoteAddress string, LastLiveTime string, CreateTime string) *Conn {
	return &Conn{RemoteAddress, LastLiveTime, CreateTime}
}
