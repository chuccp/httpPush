package auth

import "github.com/chuccp/go-web-frame/web"

const AuthKey = "wechatAuthKey"

// WithWechatAuth 标记需要进行微信授权检查的路由
func WithAuth() web.MetaOption {
	return web.WithKey(AuthKey)
}
