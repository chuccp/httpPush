package core

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"sort"
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
	userSize     int
	userIndex    int
	err          error
	hasUser      bool
	isForward    bool
}

type MsgDock struct {
	IForward   IForward
	sendQueue  *util.Queue
	replyQueue *util.Queue
	userStore  *user.Store
}

func NewMsgDock(userStore *user.Store) *MsgDock {
	msgDock := &MsgDock{sendQueue: util.NewQueue(), replyQueue: util.NewQueue(), userStore: userStore}
	msgDock.run()
	return msgDock
}
func (md *MsgDock) run() {
	go md.exchangeSendMsg()
	go md.exchangeReplyMsg()
}

func (md *MsgDock) WriteMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	username := msg.GetString(message.To)
	ius := make([]user.IOrderUser, 0)
	us, fg := md.userStore.GetOrderUser(msg.GetString(message.To))
	if fg {
		ius = append(ius, us...)
	}
	if md.IForward != nil && md.IForward.WriteMessage != nil {
		mu, fa := md.IForward.GetOrderUser(username)
		if fa {
			ius = append(ius, mu...)
		}
	}
	sort.Sort(user.ByAsc(ius))
	md.sendQueue.Offer(&DockMessage{InputMessage: msg, write: writeFunc, users: us, userIndex: -1, userSize: len(us), isForward: true})
}
func (md *MsgDock) WriteNoForwardMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	us, fg := md.userStore.GetOrderUser(msg.GetString(message.To))
	if fg {
		sort.Sort(user.ByAsc(us))
		md.sendQueue.Offer(&DockMessage{InputMessage: msg, write: writeFunc, users: us, userIndex: -1, userSize: len(us), isForward: false})
	} else {
		writeFunc(NoFoundUser, false)
	}

}
func (md *MsgDock) writeUserMsg(dockMessage *DockMessage) {
	dockMessage.userIndex++
	if (dockMessage.userIndex) < dockMessage.userSize {
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
			if md.IForward != nil && md.IForward.WriteMessage != nil {
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

	//if dockMessage.hasLocal || !dockMessage.isForward {
	//	dockMessage.userIndex++
	//	if (dockMessage.userIndex) < dockMessage.userSize {
	//		u := dockMessage.users[dockMessage.userIndex]
	//		u.WriteMessage(dockMessage.InputMessage, func(err error, hasUser bool) {
	//			if hasUser && err == nil {
	//				dockMessage.hasUser = hasUser
	//				md.replyMessage(dockMessage)
	//			} else {
	//				dockMessage.err = err
	//				md.sendQueue.Offer(dockMessage)
	//			}
	//		})
	//	} else {
	//		md.replyMessage(dockMessage)
	//	}
	//} else {
	//	if md.IForward != nil && md.IForward.WriteMessage != nil {
	//		md.IForward.WriteMessage(dockMessage.InputMessage, func(err error, hasUser bool) {
	//			dockMessage.err = err
	//			dockMessage.hasUser = hasUser
	//			md.replyMessage(dockMessage)
	//		})
	//	} else {
	//		dockMessage.hasUser = false
	//		md.replyMessage(dockMessage)
	//	}
	//}
}

func (md *MsgDock) replyMessage(msg *DockMessage) {
	md.replyQueue.Offer(msg)
}
func (md *MsgDock) HandleAddUser(iUser user.IUser) {
	if md.IForward != nil && md.IForward.HandleAddUser != nil {
		md.IForward.HandleAddUser(iUser)
	}
}

func (md *MsgDock) Query(parameter *Parameter, localValue any) []any {
	if md.IForward != nil && md.IForward.Query != nil {
		return md.IForward.Query(parameter, localValue)
	}
	return nil
}

func (md *MsgDock) HandleDeleteUser(username string) {
	if md.IForward != nil && md.IForward.HandleDeleteUser != nil {
		md.IForward.HandleDeleteUser(username)
	}
}
func (md *MsgDock) exchangeReplyMsg() {
	for {
		msg, _ := md.replyQueue.Poll()
		dockMessage := msg.(*DockMessage)
		if msg != nil {
			dockMessage.write(dockMessage.err, dockMessage.hasUser)
		}
	}
}
func (md *MsgDock) exchangeSendMsg() {
	for {
		msg, _ := md.sendQueue.Poll()
		if msg != nil {
			dm := msg.(*DockMessage)
			md.writeUserMsg(dm)
		}
	}
}
