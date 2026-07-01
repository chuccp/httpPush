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
	context   *App
}

func NewMsgDock(userStore *user.Store, context *App) *MsgDock {
	msgDock := &MsgDock{userStore: userStore, context: context}
	return msgDock
}

func (md *MsgDock) SendLocalMessage(msg message.IMessage) (bool, error) {
	username := msg.GetString(message.To)
	us := md.userStore.GetOrderUser(username)
	var err error
	var faSend = false
	for _, u := range us {
		fa, err0 := u.WriteSyncMessage(msg)
		if fa {
			faSend = true
		}
		if err0 != nil {
			err = err0
		}
	}
	if faSend {
		return true, nil
	}
	return false, err
}

func (md *MsgDock) SendMessage(msg message.IMessage) (bool, error) {
	fa, _ := md.SendLocalMessage(msg)
	if fa {
		return true, nil
	}
	if md.IForward != nil {
		return md.IForward.WriteSyncMessage(msg)
	}
	return false, NoFoundUser
}
func (md *MsgDock) Query(parameter *Parameter, localValue any) []any {
	if md.IForward != nil {
		return md.IForward.Query(parameter, localValue)
	}
	return nil
}
