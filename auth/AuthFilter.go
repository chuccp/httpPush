package auth

import (
	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/web"
)

const AuthKey = "AuthKey"

type AuthFilter struct {
	ctx *core.Context
}

func (s *AuthFilter) Init(ctx *core.Context) error {
	s.ctx = ctx
	return nil
}

func (s *AuthFilter) Handle(filterChain web.FilterChain, request *web.Request) (any, error) {
	return filterChain.Next()
}

func WithAuth() web.MetaOption {
	return web.WithKey(AuthKey)
}
