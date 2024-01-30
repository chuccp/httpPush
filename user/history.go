package user

import (
	"github.com/chuccp/httpPush/message"
	"log"
	"sync"
	"time"
)

type HistoryStore struct {
	userStore *sync.Map
}

func (h *HistoryStore) userLogin(user IUser) {
	username := user.GetUsername()
	t := time.Now()
	h.userStore.Store(username, &History{Username: username, OnlineTime: &t})
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
		h.userStore.Store(username, &History{Username: username, OnlineTime: &t, OfflineTime: &t})
	}
}
func (h *HistoryStore) RecordMessage(msg message.IMessage) {

	switch t := msg.(type) {
	case *message.TextMessage:
		{
			un := t.GetString(message.To)
			log.Println(un)

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
	}

}
func NewHistoryStore() *HistoryStore {
	return &HistoryStore{userStore: new(sync.Map)}
}

type History struct {
	Username     string
	OnlineTime   *time.Time
	OfflineTime  *time.Time
	LastLiveTime *time.Time
	LastMessage  []*TextMessage
}
type TextMessage struct {
	From string
	Msg  string
	Time *time.Time
}
