package core

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"net/http"
	"runtime/debug"
	"sync"
	"sync/atomic"
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
}

func newContext(register *Register) *Context {
	context := &Context{register: register, systemInfo: make(systemInfo)}
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

func (context *Context) RecordMessage(msg message.IMessage) {
	context.userStore.RecordMessage(msg)
}
func (context *Context) FlashLiveTime(iUser user.IUser) {
	context.userStore.FlashLiveTime(iUser)
}

func (context *Context) GetUser(userName string) ([]user.IUser, bool) {
	return context.userStore.GetUser(userName)
}
func (context *Context) GetUserCreateTime(userName string) *time.Time {
	return context.userStore.GetUserCreateTime(userName)
}
func (context *Context) GetHistory(userName string) (*user.HistoryMessage, bool) {
	return context.userStore.GetHistory(userName)
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
func (context *Context) sendOnceMessage(msg message.IMessage, write user.WriteCallBackFunc) {
	one := new(sync.Once)
	context.msgDock.WriteMessage(msg, func(err error, hasUser bool) {
		one.Do(func() {
			write(err, hasUser)
		})
	})
}

func (context *Context) SendMessageForBack(msg message.IMessage, write user.WriteCallBackFunc) {

	context.sendOnceMessage(msg, write)
}

func (context *Context) sendNoForwardOnceMessage(msg message.IMessage, write user.WriteCallBackFunc) {
	one := new(sync.Once)
	context.msgDock.WriteNoForwardMessage(msg, func(err error, hasUser bool) {
		one.Do(func() {
			write(err, hasUser)
		})
	})
}
func (context *Context) SendMessage(msg message.IMessage) (error, bool) {
	waitGroup := util.NewWaitNumGroup()
	var err_ error
	var hasUser_ = false
	waitGroup.AddOne()
	context.sendOnceMessage(msg, func(err error, hasUser bool) {
		err_ = err
		hasUser_ = hasUser
		waitGroup.Done()
	})
	waitGroup.Wait()
	return err_, hasUser_

}

func (context *Context) SendGroupTextMessage(form string, groupId, msg string) int32 {
	var num int32
	waitGroup := util.NewWaitNumGroup()
	if util.EqualsAnyIgnoreCase(groupId, "all") {
		context.userStore.RangeAllUser(func(username string) bool {
			waitGroup.AddOne()
			textMsg := message.NewTextMessage(form, username, msg)
			context.sendNoForwardOnceMessage(textMsg, func(err error, hasUser bool) {
				if hasUser {
					atomic.AddInt32(&num, 1)
				}
				waitGroup.Done()
			})
			return true
		})
	} else {
		context.userStore.RangeGroupUser(groupId, func(username string) bool {
			waitGroup.AddOne()
			textMsg := message.NewTextMessage(form, username, msg)
			context.sendNoForwardOnceMessage(textMsg, func(err error, hasUser bool) {
				if hasUser {
					atomic.AddInt32(&num, 1)
				}
				waitGroup.Done()
			})
			return true
		})
	}
	waitGroup.Wait()
	return num
}

func (context *Context) SendNoForwardMessage(msg message.IMessage) (error, bool) {
	waitGroup := util.NewWaitNumGroup()
	var err_ error
	var hasUser_ = false
	waitGroup.AddOne()
	context.sendNoForwardOnceMessage(msg, func(err error, hasUser bool) {
		err_ = err
		hasUser_ = hasUser
		waitGroup.Done()
	})
	waitGroup.Wait()
	return err_, hasUser_

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
func (context *Context) SendMultiMessageNoReplay(fromUser string, usernames []string, text string) {
	for _, v := range usernames {
		msg := message.NewTextMessage(fromUser, v, text)
		context.sendNoForwardOnceMessage(msg, func(err error, hasUser bool) {})
	}
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
