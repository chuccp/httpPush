package core

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"sort"
	"sync"
)

// IForward 集群使用/*
type IForward interface {
	HandleAddUser(iUser user.IUser)
	HandleDeleteUser(username string)
	WriteMessage(iMessage message.IMessage, exMachineId []string, writeFunc user.WriteCallBackFunc)
	GetOrderUser(username string) ([]user.IOrderUser, bool)
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
	IForward   IForward
	sendQueue  *util.Queue
	replyQueue *util.Queue
	userStore  *user.Store
	context    *Context
}

func NewMsgDock(userStore *user.Store, context *Context) *MsgDock {
	msgDock := &MsgDock{sendQueue: util.NewQueue(), replyQueue: util.NewQueue(), userStore: userStore, context: context}
	msgDock.run()
	return msgDock
}
func (md *MsgDock) run() {
	md.context.RecoverGo(func() {
		md.exchangeSendMsg()
	})
	md.context.RecoverGo(func() {
		md.exchangeReplyMsg()
	})
}

func (md *MsgDock) WriteMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	username := msg.GetString(message.To)
	ius := make([]user.IOrderUser, 0)
	us, fg := md.userStore.GetOrderUser(msg.GetString(message.To))
	if fg {
		ius = append(ius, us...)
	}
	if md.IForward != nil {
		mu, fa := md.IForward.GetOrderUser(username)
		if fa {
			ius = append(ius, mu...)
		}
	}
	if len(ius) > 0 {
		sort.Sort(user.ByAsc(ius))
	}
	//md.context.GetLog().Debug("已存在用户连接数", zap.Int("order.user.num", len(ius)))
	md.sendQueue.Offer(&DockMessage{InputMessage: msg, write: writeFunc, users: ius, userIndex: -1, isForward: true})
}
func (md *MsgDock) WriteNoForwardMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	us, fg := md.userStore.GetOrderUser(msg.GetString(message.To))
	md.context.GetLog().Debug("收到不转发信息", zap.Int("order.user.num", len(us)), zap.Bool("fa", fg))
	if fg {
		sort.Sort(user.ByAsc(us))
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
			md.replyMessage(dockMessage)
		} else {
			if md.IForward != nil {
				var exMachineIds = make([]string, 0)
				for _, orderUser := range dockMessage.users {
					if orderUser.GetMachineId() != "" {
						exMachineIds = append(exMachineIds, orderUser.GetMachineId())
					}
				}
				md.IForward.WriteMessage(dockMessage.InputMessage, exMachineIds, func(err error, hasUser bool) {
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
func (md *MsgDock) HandleAddUser(iUser user.IUser) {
	if md.IForward != nil {
		md.IForward.HandleAddUser(iUser)
	}
}

func (md *MsgDock) Query(parameter *Parameter, localValue any) []any {
	if md.IForward != nil {
		return md.IForward.Query(parameter, localValue)
	}
	return nil
}

func (md *MsgDock) HandleDeleteUser(username string) {
	if md.IForward != nil {
		md.IForward.HandleDeleteUser(username)
	}
}

func backMsg(md *MsgDock, dm *DockMessage) {
	dm.writeCallBackFunc(dm.err, dm.hasUser)
}

func (md *MsgDock) exchangeReplyMsg() {
	for {
		msg, _ := md.replyQueue.Poll()
		dockMessage := msg.(*DockMessage)
		if msg != nil {
			backMsg(md, dockMessage)
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
			dm := msg.(*DockMessage)
			sendMsg(md, dm)
		}
	}
}
