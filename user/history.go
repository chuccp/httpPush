package user

import (
	"sync"
	"time"
)

type HistoryStore struct {
	userStore *sync.Map
}

func (h *HistoryStore) userOnline(user IUser) {
	username := user.GetUsername()
	t := time.Now()
	h.userStore.Store(username, &SignUpLog{Username: username, OnlineTime: &t})
}

func (h *HistoryStore) getUserHistory(username string) (*SignUpLog, bool) {
	v, ok := h.userStore.Load(username)
	if ok {
		return v.(*SignUpLog), true
	}
	return nil, false
}

func (h *HistoryStore) userOffline(user IUser) {
	t := time.Now()
	username := user.GetUsername()
	v, ok := h.userStore.Load(username)
	if ok {
		signUpLog := v.(*SignUpLog)
		signUpLog.OfflineTime = &t
	} else {
		h.userStore.Store(username, &SignUpLog{Username: username, OnlineTime: &t, OfflineTime: &t})
	}
}

func NewHistoryStore() *HistoryStore {
	return &HistoryStore{userStore: new(sync.Map)}
}

type SignUpLog struct {
	Username    string
	OnlineTime  *time.Time
	OfflineTime *time.Time
}
