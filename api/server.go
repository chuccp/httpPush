package api

import (
	wf "github.com/chuccp/go-web-frame"
	wfcore "github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
)

type Controller struct {
	app *core.App
}

func NewController() *Controller { return &Controller{} }

func (c *Controller) Init(ctx *wfcore.Context) error {
	c.app = wf.GetService[*core.App](ctx)
	h := ctx.Get

	// 基础 API
	h("/root_version", c.rootVersion)
	h("/sendmsg", c.sendMsg)
	h("/sendMessage", c.sendMessage)
	h("/sendGroupMsg", c.sendGroupMsg)
	c.app.RegisterHandle("/sendGroupMsg", c.handleSendGroupMsg)

	// 查询 API — 同时注册端口 handler 和本地查询函数
	h("/queryUser", c.queryUser)
	c.app.RegisterHandle("/queryUser", c.handleQueryUser)
	h("/onlineUser", c.onlineUser)
	c.app.RegisterHandle("/onlineUser", c.handleOnlineUser)
	h("/info_user", c.clusterInfo)
	c.app.RegisterHandle("/info_user", c.handleClusterInfo)
	h("/queryOrderInfo", c.queryOrderInfo)
	c.app.RegisterHandle("/queryOrderInfo", c.handleQueryOrderInfo)
	h("/queryClusterUserNum", c.queryClusterUserNum)
	c.app.RegisterHandle("/queryClusterUserNum", c.handleClusterUserNum)
	h("/queryGroupInfo", c.queryGroupInfo)
	c.app.RegisterHandle("/queryGroupInfo", c.handleQueryGroupInfo)
	h("/queryVersion", c.queryVersion)
	c.app.RegisterHandle("/queryVersion", c.handleQueryVersion)

	c.app.SetSystemInfo("VERSION", core.VERSION)
	if _, ok := c.app.GetHandle("machineInfoId"); !ok {
		c.app.RegisterHandle("machineInfoId", func(p *core.Parameter) any { return "" })
		c.app.RegisterHandle("remoteMachineNum", func(p *core.Parameter) any { return 0 })
		c.app.RegisterHandle("clusterUserNum", func(p *core.Parameter) any { return 0 })
		c.app.RegisterHandle("machineAddress", func(p *core.Parameter) any { return "" })
	}

	return nil
}

func (c *Controller) rootVersion(r *web.Request) (any, error) {
	return map[string]any{
		"version":   core.VERSION,
		"startTime": c.app.GetStartTime(),
	}, nil
}

func (c *Controller) sendMsg(r *web.Request) (any, error) {
	username := r.Query("username")
	if username == "" { username = r.Query("id") }
	msg := r.Query("msg")
	if username == "" || msg == "" { return "username or msg required", nil }
	fa, _ := c.app.SendTextMessage("system", username, msg)
	if fa { return "success", nil }
	return "NO user", nil
}

func (c *Controller) sendGroupMsg(r *web.Request) (any, error) {
	groupId := r.Query("groupId")
	msg := r.Query("msg")
	if groupId == "" || msg == "" { return "groupId or msg required", nil }
	num := c.app.SendGroupTextMessage("system", groupId, msg)
	return map[string]any{"success": true, "num": num}, nil
}

func (c *Controller) handleSendGroupMsg(p *core.Parameter) any {
	return map[string]any{}
}

func (c *Controller) sendMessage(r *web.Request) (any, error) {
	username := r.Query("username")
	if username == "" { username = r.Query("id") }
	msg := r.Query("msg")
	if username == "" || msg == "" { return "username or msg required", nil }
	fa, _ := c.app.SendTextMessage("system", username, msg)
	return map[string]any{"success": fa}, nil
}

// ========== 本地查询函数 (RegisterHandle 注册，供 App.Query 调用) ==========

func (c *Controller) handleQueryUser(p *core.Parameter) any {
	id := p.GetVString("id", "username")
	var result []map[string]any
	if id != "" {
		if us, ok := c.app.GetUser(id); ok {
			for _, u := range user.SortByAsc(us) {
				result = append(result, map[string]any{
					"username":      u.GetUsername(),
					"remoteAddress": u.GetRemoteAddress(),
					"lastLiveTime":  u.LastLiveTime().Format(util.TimestampFormat),
					"createTime":    u.CreateTime().Format(util.TimestampFormat),
				})
			}
		}
	}
	return result
}

func (c *Controller) handleOnlineUser(p *core.Parameter) any {
	var list []*pageUser
	c.app.RangeUser(func(username string, _ *user.StoreUser) bool {
		list = append(list, &pageUser{UserName: username})
		return true
	})
	return &page{List: list, Num: c.app.GetUserNum()}
}

func (c *Controller) handleClusterInfo(p *core.Parameter) any {
	mid, _ := c.app.GetHandle("machineInfoId")
	machineId := ""
	if mid != nil { machineId = mid(p).(string) }
	return &machineInfo{MachineId: machineId, UserNum: c.app.GetUserNum()}
}

func (c *Controller) handleQueryOrderInfo(p *core.Parameter) any {
	userId := p.GetVString("userId", "username", "id")
	us := c.app.GetUserOrder(userId)
	var list []*orderUser
	for _, u := range us {
		list = append(list, &orderUser{Priority: u.GetPriority(), MachineId: u.GetMachineId(), OrderTime: u.GetOrderTime().Format(util.TimestampFormat)})
	}
	return &allOrderUser{OrderUser: list}
}

func (c *Controller) handleClusterUserNum(p *core.Parameter) any {
	h, ok := c.app.GetHandle("clusterUserNum")
	if ok { return &clusterUserNum{UserNum: h(p)} }
	return &clusterUserNum{}
}

func (c *Controller) handleQueryGroupInfo(p *core.Parameter) any {
	return &groupInfo{GroupInfo: c.app.AllGroupInfo()}
}

func (c *Controller) handleQueryVersion(p *core.Parameter) any {
	return &versionInfo{Version: core.VERSION, StartTime: c.app.GetStartTime()}
}

// ========== API handler ==========

func (c *Controller) queryUser(r *web.Request) (any, error) {
	result := make([]any, 0)
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	for _, v := range vs {
		result = append(result, v)
	}
	return result, nil
}

func (c *Controller) onlineUser(r *web.Request) (any, error) {
	parameter := newParameter(r)
	values := c.app.Query(parameter).([]any)
	result := make([]map[string]any, 0)
	for _, v := range values {
		if p, ok := v.(*page); ok {
			for _, u := range p.List {
				result = append(result, map[string]any{
					"userName":   u.UserName,
					"machineId":  u.MachineId,
					"createTime": u.CreateTime,
				})
			}
		}
	}
	if len(result) == 0 {
		c.app.RangeUser(func(username string, _ *user.StoreUser) bool {
			result = append(result, map[string]any{"userName": username})
			return true
		})
	}
	return result, nil
}

type machineInfo struct {
	Address   string `json:"address,omitempty"`
	UserNum   int    `json:"userNum"`
	MachineId string `json:"machineId"`
}
type pageUser struct {
	UserName       string
	MachineAddress string
	CreateTime     string
	MachineId      string
}
type page struct {
	Num  int
	List []*pageUser
}
type clusterUserNum struct {
	UserNum   any    `json:"userNum"`
	MachineId string `json:"machineId"`
}
type groupInfo struct {
	MachineId string
	GroupInfo map[string]int
}
type versionInfo struct {
	Version   string
	StartTime string
	MachineId string
}
type orderUser struct {
	Priority  int
	MachineId string
	OrderTime string
}
type allOrderUser struct {
	OrderUser []*orderUser
	MachineId string
}

func (c *Controller) clusterInfo(r *web.Request) (any, error) {
	parameter := newParameter(r)
	values := c.app.Query(parameter).([]any)
	result := make([]*machineInfo, 0)
	total := 0
	for _, value := range values {
		if mi, ok := value.(*machineInfo); ok {
			result = append(result, mi)
			total += mi.UserNum
		}
	}
	return map[string]any{"cluster": result, "total": total}, nil
}

func (c *Controller) queryOrderInfo(r *web.Request) (any, error) {
	parameter := newParameter(r)
	values := c.app.Query(parameter).([]any)
	result := make([]any, 0)
	for _, v := range values {
		if ao, ok := v.(*allOrderUser); ok {
			for _, u := range ao.OrderUser {
				result = append(result, map[string]any{
					"priority": u.Priority, "machineId": u.MachineId, "orderTime": u.OrderTime,
					"machineAddress": ao.MachineId,
				})
			}
		}
	}
	// fallback: just show local
	if len(result) == 0 {
		userId := r.Query("id")
		if userId == "" { userId = r.Query("username") }
		for _, u := range c.app.GetUserOrder(userId) {
			result = append(result, map[string]any{
				"priority": u.GetPriority(), "machineId": u.GetMachineId(),
				"orderTime": u.GetOrderTime().Format(util.TimestampFormat),
			})
		}
	}
	return result, nil
}

func (c *Controller) queryClusterUserNum(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	result := make([]*clusterUserNum, 0)
	for _, v := range vs {
		if m, ok := v.(*clusterUserNum); ok { result = append(result, m) }
	}
	return result, nil
}

func (c *Controller) queryGroupInfo(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	result := make([]*groupInfo, 0)
	for _, v := range vs {
		if m, ok := v.(*groupInfo); ok { result = append(result, m) }
	}
	return result, nil
}

func (c *Controller) queryVersion(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	result := make([]*versionInfo, 0)
	for _, v := range vs {
		if m, ok := v.(*versionInfo); ok { result = append(result, m) }
	}
	return map[string]any{
		"versions": result,
		"local":    map[string]string{"version": core.VERSION, "startTime": c.app.GetStartTime()},
	}, nil
}

func newParameter(r *web.Request) *core.Parameter {
	return &core.Parameter{
		Path:    r.FullPath(),
		Form:    extractValues(r),
		SetFrom: make(map[string][]string),
	}
}

func extractValues(r *web.Request) map[string][]string {
	result := make(map[string][]string)
	for k, v := range r.GinContext().Request.URL.Query() {
		result[k] = v
	}
	return result
}
