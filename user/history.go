package user

import (
	"github.com/chuccp/httpPush/message"
	"sync"
	"time"
)

type HistoryStore struct {
	userStore *sync.Map
}

func (h *HistoryStore) userLogin(user IUser) {
	username := user.GetUsername()
	h.userStore.Store(username, newHistory(username))
}

func (h *HistoryStore) getUserHistory(username string) (*History, bool) {
	v, ok := h.userStore.Load(username)
	if ok {
		return v.(*History), true
	}
	return nil, false
}

func (h *HistoryStore) userOffline(user IUser) {
	t := time.Now()
	username := user.GetUsername()
	v, ok := h.userStore.Load(username)
	if ok {
		signUpLog := v.(*History)
		signUpLog.OfflineTime = &t
	} else {
		h.userStore.Store(username, newHistory(username))
	}
}
func (h *HistoryStore) RecordMessage(msg message.IMessage) {

	switch mg := msg.(type) {
	case *message.TextMessage:
		{
			username := mg.GetString(message.To)
			v, ok := h.userStore.Load(username)
			if ok {
				history := v.(*History)
				history.recordMessage(mg)
			} else {
				history := newHistory(username)
				history.recordMessage(mg)
				h.userStore.Store(username, history)
			}
		}
	}
}
func (h *HistoryStore) FlashLiveTime(user IUser) {
	t := time.Now()
	username := user.GetUsername()
	v, ok := h.userStore.Load(username)
	if ok {
		signUpLog := v.(*History)
		signUpLog.LastLiveTime = &t
	} else {
		username := user.GetUsername()
		h.userStore.Store(username, newHistory(username))
	}
}
func NewHistoryStore() *HistoryStore {
	return &HistoryStore{userStore: new(sync.Map)}
}

func newHistory(username string) *History {
	t := time.Now()
	history := &History{Username: username, OnlineTime: &t, LastLiveTime: &t, LastMessage: make([]*TextMessage, 0), rLock: new(sync.RWMutex)}
	return history
}

type History struct {
	Username     string
	OnlineTime   *time.Time
	OfflineTime  *time.Time
	LastLiveTime *time.Time
	LastMessage  []*TextMessage
	rLock        *sync.RWMutex
}

func (h *History) recordMessage(msg *message.TextMessage) {
	h.rLock.Lock()
	defer h.rLock.Unlock()
	if len(h.LastMessage) > 5 {
		h.LastMessage = h.LastMessage[1:]
	}
	t := time.Now()
	h.LastMessage = append(h.LastMessage, &TextMessage{From: msg.From, Msg: msg.Msg, Time: &t})
}
func (h *History) readMessage() []*TextMessage {
	h.rLock.RLock()
	defer h.rLock.RUnlock()
	return h.LastMessage
}

type TextMessage struct {
	From string
	Msg  string
	Time *time.Time
}
