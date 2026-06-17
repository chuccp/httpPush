package core

import (
	"time"

	"github.com/chuccp/go-web-frame/config"
	wfcore "github.com/chuccp/go-web-frame/core"
	wflog "github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"go.uber.org/zap"
)

const VERSION = "0.5.0"

// Context 向后兼容别名
type Context = App

// App httpPush 全局共享状态，实现 go-web-frame IService
type App struct {
	userStore     *user.Store
	msgDock       *MsgDock
	handleFuncMap map[string]RegisterHandle
	startTime     *time.Time
	cfg           config.IConfig
	systemInfo    map[string]any
}

func NewApp() *App {
	st := time.Now()
	app := &App{
		startTime:     &st,
		handleFuncMap: make(map[string]RegisterHandle),
		systemInfo:    make(map[string]any),
	}
	app.userStore = user.NewStore()
	app.msgDock = NewMsgDock(app.userStore, app)
	return app
}

func (a *App) Init(ctx *wfcore.Context) error {
	a.cfg = ctx.GetConfig()
	return nil
}

func (a *App) GetStartTime() string           { return a.startTime.Format("2006-01-02 15:04:05.000") }
func (a *App) SetForward(forward IForward)     { a.msgDock.IForward = forward }

func (a *App) AddUser(iUser user.IUser) {
	if a.userStore.AddUser(iUser) {
		wflog.Info("新增用户", zap.String("username", iUser.GetUsername()))
	}
}
func (a *App) GetUserOrder(username string) []user.IOrderUser { return a.userStore.GetOrderUser(username) }
func (a *App) GetUser(userName string) ([]user.IUser, bool)   { return a.userStore.GetUser(userName) }
func (a *App) GetUserCreateTime(userName string) *time.Time    { return a.userStore.GetUserCreateTime(userName) }
func (a *App) GetUserNum() int                                 { return a.userStore.GetUserNum() }
func (a *App) DeleteUser(iUser user.IUser) bool {
	if a.userStore.DeleteUser(iUser) {
		wflog.Info("用户离线", zap.String("username", iUser.GetUsername()))
		return true
	}
	return false
}
func (a *App) RangeUser(f func(username string, user *user.StoreUser) bool) { a.userStore.Range(f) }
func (a *App) AllGroupInfo() map[string]int                                  { return a.userStore.AllGroupInfo() }
func (a *App) HasLocalUser(username string) bool                             { return a.userStore.HasLocalUser(username) }

func (a *App) SendLocalMessage(msg message.IMessage) (bool, error) { return a.msgDock.SendLocalMessage(msg) }
func (a *App) SendMessage(msg message.IMessage) (bool, error)      { return a.msgDock.SendMessage(msg) }
func (a *App) SendTextMessage(from, to, msg string) (bool, error) {
	return a.SendMessage(message.NewTextMessage(from, to, msg))
}
func (a *App) SendGroupTextMessage(from, groupId, msg string) int32 {
	var num int32
	if groupId == "all" || groupId == "All" {
		a.userStore.RangeAllUser(func(username string) bool {
			if fa, _ := a.msgDock.SendLocalMessage(message.NewTextMessage(from, username, msg)); fa { num++ }
			return true
		})
	} else {
		a.userStore.RangeGroupUser(groupId, func(username string) bool {
			if fa, _ := a.msgDock.SendLocalMessage(message.NewTextMessage(from, username, msg)); fa { num++ }
			return true
		})
	}
	return num
}

func (a *App) Query(parameter *Parameter) any {
	iv := make([]any, 0)
	if h, ok := a.GetHandle(parameter.Path); ok {
		v := h(parameter)
		iv = append(iv, v)
		if vs := a.msgDock.Query(parameter, v); vs != nil {
			iv = append(iv, vs...)
		}
	}
	return iv
}

func (a *App) RegisterHandle(name string, h RegisterHandle) { a.handleFuncMap[name] = h }
func (a *App) GetHandle(name string) (RegisterHandle, bool)  { h, ok := a.handleFuncMap[name]; return h, ok }
func (a *App) GetSystemInfo() map[string]any                 { return a.systemInfo }
func (a *App) SetSystemInfo(key string, value any)           { a.systemInfo[key] = value }

// Config helpers
func (a *App) GetCfgString(section, key string) string            { return a.cfg.GetString(section + "." + key) }
func (a *App) GetCfgInt(section, key string) int                  { return a.cfg.GetInt(section + "." + key) }
func (a *App) GetCfgBoolDefault(section, key string, d bool) bool { return a.cfg.GetBoolOrDefault(section+"."+key, d) }
func (a *App) GetCfgStringDefault(section, key, d string) string  { return a.cfg.GetStringOrDefault(section+"."+key, d) }
