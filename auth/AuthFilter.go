package auth

import (
	"errors"

	"github.com/chuccp/go-web-frame/core"
	"github.com/chuccp/go-web-frame/log"
	"github.com/chuccp/go-web-frame/web"
)

type AuthFilter struct {
	ctx   *core.Context
	token string
}

func NewAuthFilter() *AuthFilter {
	return &AuthFilter{}
}

func (s *AuthFilter) Init(ctx *core.Context) error {
	s.ctx = ctx
	s.token = ctx.GetConfig().GetStringOrDefault("auth.token", "")
	if s.token == "" {
		log.Warn("auth.token is not configured, authentication will always fail")
	}
	return nil
}

func (s *AuthFilter) Handle(filterChain web.FilterChain, request *web.Request) (any, error) {
	if len(s.token) > 0 {
		if request.HandlerMeta().Has(Key) {
			token := s.extractToken(request)
			if token != s.token {
				return nil, errors.New("invalid or missing token")
			}
		}
	}
	return filterChain.Next()
}

// extractToken tries to get the token from query params, JSON body, or form params.
func (s *AuthFilter) extractToken(request *web.Request) string {
	if token := request.Query("token"); token != "" {
		return token
	}
	if token, err := request.GetJsonStringValue("token"); err == nil && token != "" {
		return token
	}
	if token := request.GetFormParam("token"); token != "" {
		return token
	}
	return ""
}
