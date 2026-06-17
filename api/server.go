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
	h := ctx.Get  // route register shorthand

	// 基础 API
	h("/root_version", c.rootVersion)
	h("/sendmsg", c.sendMsg)
	h("/sendMessage", c.sendMessage)

	// 查询 API
	h("/queryUser", c.queryUser)
	h("/onlineUser", c.onlineUser)
	h("/info_user", c.clusterInfo)
	h("/queryOrderInfo", c.queryOrderInfo)
	h("/queryClusterUserNum", c.queryClusterUserNum)
	h("/queryGroupInfo", c.queryGroupInfo)
	h("/queryVersion", c.queryVersion)

	c.app.SetSystemInfo("VERSION", core.VERSION)
	// 只在 cluster 未注册时才设默认桩
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
	if username == "" {
		username = r.Query("id")
	}
	msg := r.Query("msg")
	if username == "" || msg == "" {
		return "username or msg required", nil
	}
	fa, _ := c.app.SendTextMessage("system", username, msg)
	if fa {
		return "success", nil
	}
	return "NO user", nil
}

func (c *Controller) sendMessage(r *web.Request) (any, error) {
	username := r.Query("username")
	if username == "" {
		username = r.Query("id")
	}
	msg := r.Query("msg")
	if username == "" || msg == "" {
		return "username or msg required", nil
	}
	fa, _ := c.app.SendTextMessage("system", username, msg)
	return map[string]any{"success": fa}, nil
}

func (c *Controller) queryUser(r *web.Request) (any, error) {
	id := r.Query("id")
	if id == "" {
		id = r.Query("username")
	}
	result := make([]any, 0)
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	for _, v := range vs {
		result = append(result, v)
	}
	if id != "" {
		if us, ok := c.app.GetUser(id); ok {
			sorted := user.SortByAsc(us)
			for _, u := range sorted {
				result = append(result, map[string]any{
					"username":      u.GetUsername(),
					"remoteAddress": u.GetRemoteAddress(),
					"lastLiveTime":  u.LastLiveTime().Format(util.TimestampFormat),
					"createTime":    u.CreateTime().Format(util.TimestampFormat),
				})
			}
		}
	}
	return result, nil
}

func (c *Controller) onlineUser(r *web.Request) (any, error) {
	parameter := newParameter(r)
	values := c.app.Query(parameter).([]any)
	result := make([]map[string]any, 0)
	for _, v := range values {
		// 本地返回的是 *page，远程 unmarshal 后也是 *page
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

// 与集群 query 返回类型匹配的结构体
type pageUser struct {
	UserName       string
	MachineAddress string
	CreateTime     string
	MachineId      string
}
type page struct {
	Num        int
	List       []*pageUser
	MachineId  string
	UserNum    int
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

func (c *Controller) clusterInfo(r *web.Request) (any, error) {
	parameter := newParameter(r)
	values := c.app.Query(parameter).([]any)
	result := make([]*machineInfo, 0)
	total := 0
	for _, value := range values {
		mi, ok := value.(*machineInfo)
		if !ok {
			// 远程返回的可能是 any 包装
			if a, ok2 := value.(*any); ok2 {
				mi, _ = (*a).(*machineInfo)
			}
		}
		if mi != nil {
			result = append(result, mi)
			total += mi.UserNum
		}
	}
	return map[string]any{
		"cluster": result,
		"total":   total,
	}, nil
}

func (c *Controller) queryOrderInfo(r *web.Request) (any, error) {
	userId := r.Query("id")
	if userId == "" {
		userId = r.Query("username")
	}
	us := c.app.GetUserOrder(userId)
	result := make([]map[string]any, 0)
	for _, u := range us {
		result = append(result, map[string]any{
			"priority":  u.GetPriority(),
			"machineId": u.GetMachineId(),
			"orderTime": u.GetOrderTime().Format(util.TimestampFormat),
		})
	}
	return result, nil
}

func (c *Controller) queryClusterUserNum(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	result := make([]*clusterUserNum, 0)
	for _, v := range vs {
		if m, ok := v.(*clusterUserNum); ok {
			result = append(result, m)
		}
	}
	return result, nil
}

func (c *Controller) queryGroupInfo(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	result := make([]*groupInfo, 0)
	for _, v := range vs {
		if m, ok := v.(*groupInfo); ok {
			result = append(result, m)
		}
	}
	return result, nil
}

func (c *Controller) queryVersion(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	result := make([]*versionInfo, 0)
	for _, v := range vs {
		if m, ok := v.(*versionInfo); ok {
			result = append(result, m)
		}
	}
	return map[string]any{
		"versions": result,
		"local": map[string]string{
			"version":   core.VERSION,
			"startTime": c.app.GetStartTime(),
		},
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
