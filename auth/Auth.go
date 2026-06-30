package auth

import (
	"github.com/chuccp/go-web-frame/web"
)

const Key = "AuthKey"

func WithAuth() web.MetaOption {
	return web.WithKey(Key)
}
