package core

import (
	"errors"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"sync"
)

// IForward 集群使用/*
type IForward interface {
	WriteMessage(iMessage message.IMessage, writeFunc user.WriteCallBackFunc)
	Query(parameter *Parameter, localValue any) []any
}

type DockMessage struct {
	InputMessage message.IMessage
	write        user.WriteCallBackFunc
	users        []user.IOrderUser
	userIndex    int
	err          error
	hasUser      bool
	isForward    bool
	once         sync.Once
}

func (m *DockMessage) writeCallBackFunc(err error, hasUser bool) {
	m.once.Do(func() {
		m.write(err, hasUser)
	})
}

type MsgDock struct {
	IForward             IForward
	sendQueue            *util.Queue
	replyQueue           *util.Queue
	userStore            *user.Store
	context              *Context
	lastSendDockMessage  *DockMessage
	lastReplyDockMessage *DockMessage
}

func NewMsgDock(userStore *user.Store, context *Context) *MsgDock {
	msgDock := &MsgDock{sendQueue: util.NewQueue(), replyQueue: util.NewQueue(), userStore: userStore, context: context}
	msgDock.run()
	return msgDock
}
func (md *MsgDock) run() {
	md.context.RecoverGo(func() {
		if md.lastSendDockMessage != nil {
			lastSendDockMessage := md.lastSendDockMessage
			md.lastSendDockMessage = nil
			lastSendDockMessage.writeCallBackFunc(errors.New("系统异常"), false)

		}
		md.exchangeSendMsg()
	})
	md.context.RecoverGo(func() {
		if md.lastReplyDockMessage != nil {
			lastReplyDockMessage := md.lastReplyDockMessage
			md.lastReplyDockMessage = nil
			lastReplyDockMessage.writeCallBackFunc(errors.New("系统异常"), false)
		}
		md.exchangeReplyMsg()
	})
}
func (md *MsgDock) WriteMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	username := msg.GetString(message.To)
	us := md.userStore.GetOrderUser(username)
	md.sendQueue.Offer(&DockMessage{InputMessage: msg, write: writeFunc, users: us, userIndex: -1, isForward: true})
}
func (md *MsgDock) WriteNoForwardMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	us := md.userStore.GetOrderUser(msg.GetString(message.To))
	md.context.GetLog().Debug("收到不转发信息", zap.Int("order.user.num", len(us)), zap.Bool("fa", len(us) > 0))
	if len(us) > 0 {
		num := md.sendQueue.Offer(&DockMessage{InputMessage: msg, write: writeFunc, users: us, userIndex: -1, isForward: false})
		md.context.GetLog().Debug("收到不转发信息入库", zap.Int("dockMessage.userIndex", int(num)))
	} else {
		writeFunc(NoFoundUser, false)
	}

}
func (md *MsgDock) writeUserMsg(dockMessage *DockMessage) {
	dockMessage.userIndex++
	if (dockMessage.userIndex) < len(dockMessage.users) {
		u := dockMessage.users[dockMessage.userIndex]
		u.WriteMessage(dockMessage.InputMessage, func(err error, hasUser bool) {
			if hasUser && err == nil {
				dockMessage.hasUser = hasUser
				md.replyMessage(dockMessage)
			} else {
				dockMessage.err = err
				md.sendQueue.Offer(dockMessage)
			}
		})
	} else {
		if !dockMessage.isForward {
			dockMessage.hasUser = false
			md.replyMessage(dockMessage)
		} else {
			if md.IForward != nil {
				md.IForward.WriteMessage(dockMessage.InputMessage, func(err error, hasUser bool) {
					dockMessage.err = err
					dockMessage.hasUser = hasUser
					md.replyMessage(dockMessage)
				})
			} else {
				dockMessage.hasUser = false
				md.replyMessage(dockMessage)
			}

		}
	}
}

func (md *MsgDock) replyMessage(msg *DockMessage) {
	md.replyQueue.Offer(msg)
}
func (md *MsgDock) Query(parameter *Parameter, localValue any) []any {
	if md.IForward != nil {
		return md.IForward.Query(parameter, localValue)
	}
	return nil
}
func backMsg(md *MsgDock, dm *DockMessage) {
	dm.writeCallBackFunc(dm.err, dm.hasUser)
}

func (md *MsgDock) exchangeReplyMsg() {
	for {
		msg, _ := md.replyQueue.Poll()
		md.lastReplyDockMessage = msg.(*DockMessage)
		if msg != nil {
			backMsg(md, md.lastReplyDockMessage)
		}
	}
}
func sendMsg(md *MsgDock, dm *DockMessage) {
	md.writeUserMsg(dm)
}

func (md *MsgDock) exchangeSendMsg() {
	for {
		msg, _ := md.sendQueue.Poll()
		if msg != nil {
			md.lastSendDockMessage = msg.(*DockMessage)
			sendMsg(md, md.lastSendDockMessage)
		}
	}
}
