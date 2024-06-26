package core

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
)

// IForward 集群使用/*
type IForward interface {
	WriteSyncMessage(iMessage message.IMessage) (bool, error)
	Query(parameter *Parameter, localValue any) []any
}

type MsgDock struct {
	IForward  IForward
	userStore *user.Store
	context   *Context
}

func NewMsgDock(userStore *user.Store, context *Context) *MsgDock {
	msgDock := &MsgDock{userStore: userStore, context: context}
	return msgDock
}

func (md *MsgDock) SendLocalMessage(msg message.IMessage) (bool, error) {
	username := msg.GetString(message.To)
	us := md.userStore.GetOrderUser(username)
	var err error
	var fa bool
	for _, u := range us {
		fa, err = u.WriteSyncMessage(msg)
		if fa {
			return true, nil
		}
	}
	return fa, err
}

func (md *MsgDock) SendMessage(msg message.IMessage) (bool, error) {
	fa, _ := md.SendLocalMessage(msg)
	if fa {
		return true, nil
	}
	return md.IForward.WriteSyncMessage(msg)
}
func (md *MsgDock) Query(parameter *Parameter, localValue any) []any {
	if md.IForward != nil {
		return md.IForward.Query(parameter, localValue)
	}
	return nil
}
