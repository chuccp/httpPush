package core

type Server interface {
	Start() error
	Init(context *Context)
	Name() string
}
