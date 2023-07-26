package core

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"log"
	"net/http"
	"sync"
)

type systemInfo map[string]any

type Context struct {
	register      *Register
	systemInfo    systemInfo
	httpPush      *HttpPush
	userStore     *user.Store
	msgDock       *MsgDock
	handleFuncMap map[string]RegisterHandle
}

func newContext(register *Register) *Context {
	context := &Context{register: register, systemInfo: make(systemInfo)}
	context.httpPush = newHttpPush(context)
	context.userStore = user.NewStore()
	context.msgDock = NewMsgDock(context.userStore)
	context.handleFuncMap = make(map[string]RegisterHandle)
	return context
}
func (context *Context) GetHttpPush() *HttpPush {
	return context.httpPush
}
func (context *Context) SetForward(forward IForward) {
	context.msgDock.IForward = forward
}
func (context *Context) AddUser(iUser user.IUser) {
	if context.userStore.AddUser(iUser) {
		context.msgDock.HandleAddUser(iUser)
	}
}

func (context *Context) GetUser(userName string) ([]user.IUser, bool) {
	return context.userStore.GetUser(userName)
}

func (context *Context) DeleteUser(iUser user.IUser) {
	if context.userStore.DeleteUser(iUser) {
		context.msgDock.HandleAddUser(iUser)
	}
}
func (context *Context) sendMessage(msg message.IMessage, write user.WriteCallBackFunc) {
	context.msgDock.WriteMessage(msg, write)
}
func (context *Context) sendNoForwardMessage(msg message.IMessage, write user.WriteCallBackFunc) {
	context.msgDock.WriteNoForwardMessage(msg, write)
}
func (context *Context) SendMessage(msg message.IMessage) (error, bool) {
	var once sync.Once
	flag := util.GetChanBool()
	var err_ error
	context.sendMessage(msg, func(err error, hasUser bool) {
		log.Println("zzzzz", err, hasUser)
		err_ = err
		once.Do(func() {
			flag <- hasUser
		})
	})
	fg := <-flag
	util.FreeChanBool(flag)
	return err_, fg

}
func (context *Context) SendNoForwardMessage(msg message.IMessage) (error, bool) {
	var once sync.Once
	flag := util.GetChanBool()
	var err_ error
	context.sendNoForwardMessage(msg, func(err error, hasUser bool) {
		err_ = err
		once.Do(func() {
			flag <- hasUser
		})
	})
	fg := <-flag
	util.FreeChanBool(flag)
	return err_, fg

}
func (context *Context) SendTextMessage(from string, to string, msg string) (error, bool) {
	textMsg := message.NewTextMessage(from, to, msg)
	return context.SendMessage(textMsg)
}

func (context *Context) Query(parameter *Parameter) any {
	iv := make([]any, 0)
	registerHandle, fa := context.GetHandle(parameter.Path)
	log.Println(parameter.Path, registerHandle, fa)
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
func (context *Context) GetSystemInfo() systemInfo {
	return context.systemInfo
}
func (context *Context) SetSystemInfo(key string, value any) {
	context.systemInfo[key] = value
}
func (context *Context) GetCfgInt(section, key string) int {
	return context.register.config.GetInt(section, key)
}
func (context *Context) AddHttpRoute(pattern string, handler func(http.ResponseWriter, *http.Request)) {
	context.httpPush.httpServer.AddRoute(pattern, handler)
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