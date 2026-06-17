package api

import (
	"encoding/json"

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
	vs := c.app.Query(parameter).([]any)
	type item struct {
		UserName   string `json:"userName"`
		MachineId  string `json:"machineId"`
		CreateTime string `json:"createTime"`
		Conn       []any  `json:"conn"`
	}
	result := make([]item, 0)
	for _, v := range vs {
		if p, ok := v.(map[string]any); ok {
			result = append(result, item{
				UserName:   p["userName"].(string),
				MachineId:  p["machineId"].(string),
				CreateTime: p["createTime"].(string),
			})
		}
	}
	// fallback: just list local users
	if len(result) == 0 {
		c.app.RangeUser(func(username string, _ *user.StoreUser) bool {
			result = append(result, item{UserName: username})
			return true
		})
	}
	return result, nil
}

func (c *Controller) clusterInfo(r *web.Request) (any, error) {
	parameter := newParameter(r)
	values := c.app.Query(parameter).([]any)
	result := make([]map[string]any, 0)
	total := 0
	for _, v := range values {
		if m, ok := v.(map[string]any); ok {
			total += int(m["userNum"].(float64))
			result = append(result, m)
		}
	}
	machineId := ""
	if handle, ok := c.app.GetHandle("machineInfoId"); ok {
		machineId = handle(nil).(string)
	}
	return map[string]any{
		"total":   total + c.app.GetUserNum(),
		"cluster": append(result, map[string]any{
			"machineId": machineId,
			"userNum":   c.app.GetUserNum(),
		}),
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
	type item struct {
		MachineId string `json:"machineId"`
		UserNum   any    `json:"userNum"`
	}
	result := make([]item, 0)
	for _, v := range vs {
		if m, ok := v.(map[string]any); ok {
			result = append(result, item{MachineId: m["machineId"].(string), UserNum: m["userNum"]})
		}
	}
	return result, nil
}

func (c *Controller) queryGroupInfo(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	data, _ := json.Marshal(vs)
	return string(data), nil
}

func (c *Controller) queryVersion(r *web.Request) (any, error) {
	parameter := newParameter(r)
	vs := c.app.Query(parameter).([]any)
	return map[string]any{
		"versions": vs,
		"local":    core.VERSION,
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
