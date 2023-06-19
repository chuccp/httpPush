package core

import (
	"net/http"
)

type systemInfo map[string]any

type Context struct {
	register   *Register
	systemInfo systemInfo
	httpPush   *HttpPush
}

func newContext(register *Register) *Context {
	context := &Context{register: register, systemInfo: make(systemInfo)}
	context.httpPush = newHttpPush(context)
	return context
}
func (context *Context) GetHttpPush() *HttpPush {
	return context.httpPush
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
