package core

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

type systemInfo map[string]any

type Context struct {
	register      *Register
	systemInfo    systemInfo
	httpPush      *HttpPush
	userStore     *user.Store
	msgDock       *MsgDock
	handleFuncMap map[string]RegisterHandle
	log           *zap.Logger
	sendPool      *ants.Pool
}

func newContext(register *Register) *Context {
	pool, _ := ants.NewPool(50)
	context := &Context{register: register, systemInfo: make(systemInfo)}
	context.sendPool = pool
	context.httpPush = newHttpPush(context)
	context.userStore = user.NewStore()
	context.msgDock = NewMsgDock(context.userStore, context)
	context.handleFuncMap = make(map[string]RegisterHandle)
	return context
}
func (context *Context) GetHttpPush() *HttpPush {
	return context.httpPush
}

// RecoverGo 协程异常恢复，异常后，会重启当前协程
func (context *Context) RecoverGo(handle func()) {
	go func() {
		wg := new(sync.WaitGroup)
		for {
			wg.Add(1)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						s := string(debug.Stack())
						context.GetLog().Error("recoverGo", zap.Any("err", err), zap.String("info", s))
						wg.Done()
					}
				}()
				handle()
			}()
			wg.Wait()
		}
	}()
}

// Go 协程异常保活，避免协程内错误导致整个系统
func (context *Context) Go(handle func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s := string(debug.Stack())
				context.GetLog().Error("Go", zap.Any("err", err), zap.String("info", s))
			}
		}()
		handle()
	}()
}

// GoForIndex  协程异常保活，避免协程内错误导致整个系统
func (context *Context) GoForIndex(index int, handle func(ind int)) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				s := string(debug.Stack())
				context.GetLog().Error("Go", zap.Any("err", err), zap.String("info", s))
			}
		}()
		handle(index)
	}()
}

func (context *Context) GetUserOrder(username string) []user.IOrderUser {
	return context.userStore.GetOrderUser(username)
}
func (context *Context) GetLog() *zap.Logger {
	return context.log
}
func (context *Context) SetForward(forward IForward) {
	context.msgDock.IForward = forward
}
func (context *Context) AddUser(iUser user.IUser) {
	if context.userStore.AddUser(iUser) {
		context.log.Info("新增用户", zap.String("username", iUser.GetUsername()), zap.String("remoteAddress", iUser.GetRemoteAddress()))
	}
}

func (context *Context) GetUser(userName string) ([]user.IUser, bool) {
	return context.userStore.GetUser(userName)
}
func (context *Context) GetUserCreateTime(userName string) *time.Time {
	return context.userStore.GetUserCreateTime(userName)
}
func (context *Context) GetUserNum() int {
	return context.userStore.GetUserNum()
}
func (context *Context) RangeUser(f func(username string, user *user.StoreUser) bool) {
	context.userStore.Range(f)
}

func (context *Context) DeleteUser(iUser user.IUser) bool {
	if context.userStore.DeleteUser(iUser) {
		context.log.Info("用户离线", zap.String("username", iUser.GetUsername()), zap.String("remoteAddress", iUser.GetRemoteAddress()))
		return true
	}
	return false
}

func (context *Context) SendLocalMessage(msg message.IMessage) (err error, fa bool) {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	context.sendPool.Submit(func() {
		fa, err = context.msgDock.WriteLocalMessage(msg)
		waitGroup.Done()
	})
	waitGroup.Wait()
	return nil, false
}

func (context *Context) SendMessage(msg message.IMessage) (err error, fa bool) {
	waitGroup := new(sync.WaitGroup)
	waitGroup.Add(1)
	context.sendPool.Submit(func() {
		fa, err = context.msgDock.SendMessage(msg)
		waitGroup.Done()
	})
	waitGroup.Wait()
	return
}

func (context *Context) SendGroupTextMessage(form string, groupId, msg string) int32 {
	var num int32
	if util.EqualsAnyIgnoreCase(groupId, "all") {
		context.userStore.RangeAllUser(func(username string) bool {
			textMsg := message.NewTextMessage(form, username, msg)
			fa, _ := context.msgDock.WriteLocalMessage(textMsg)
			if fa {
				num++
			}
			return true
		})
	} else {
		context.userStore.RangeGroupUser(groupId, func(username string) bool {
			textMsg := message.NewTextMessage(form, username, msg)
			fa, _ := context.msgDock.WriteLocalMessage(textMsg)
			if fa {
				num++
			}
			return true
		})
	}
	return num
}

func (context *Context) SendTextMessage(from string, to string, msg string) (error, bool) {
	textMsg := message.NewTextMessage(from, to, msg)
	return context.SendMessage(textMsg)
}
func (context *Context) Query(parameter *Parameter) any {
	iv := make([]any, 0)
	registerHandle, fa := context.GetHandle(parameter.Path)
	if fa {
		v := registerHandle(parameter)
		iv = append(iv, v)
		vs := context.msgDock.Query(parameter, v)
		if vs != nil {
			for _, v := range vs {
				iv = append(iv, v)
			}
		}
	}
	return iv
}

func (context *Context) GetCfgString(section, key string) string {
	iSection := context.register.config.GetString(section, key)
	return iSection
}

func (context *Context) GetCfgStringDefault(section, key, defaultValue string) string {
	value := context.GetCfgString(section, key)
	if len(value) == 0 {
		return defaultValue
	}
	return value
}

func (context *Context) GetSystemInfo() systemInfo {
	return context.systemInfo
}
func (context *Context) SetSystemInfo(key string, value any) {
	context.systemInfo[key] = value
}
func (context *Context) GetCfgInt(section, key string) int {
	return context.register.config.GetInt(section, key)
}

func (context *Context) GetCfgBool(section, key string) bool {
	return context.register.config.GetBool(section, key)
}
func (context *Context) GetCfgBoolDefault(section, key string, defaultValue bool) bool {
	return context.register.config.GetBoolOrDefault(section, key, defaultValue)
}

func (context *Context) addHttpRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	context.httpPush.httpServer.AddRoute(pattern, handler)
}
func (context *Context) isTls() bool {
	return context.httpPush.httpServer.IsTls()
}
func (context *Context) rangeServer(f func(server Server)) {
	context.register.rangeServer(f)
}

func (context *Context) RegisterHandle(handleName string, handle RegisterHandle) {
	context.handleFuncMap[handleName] = handle
}

func (context *Context) GetHandle(handleName string) (RegisterHandle, bool) {
	v, ok := context.handleFuncMap[handleName]
	if ok {
		return v, true
	} else {
		return nil, false
	}
}
