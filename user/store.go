package user

import (
	"github.com/chuccp/httpPush/util"
	"sync"
	"sync/atomic"
	"time"
)

type StoreUser struct {
	store      map[string]IUser
	username   string
	createTime *time.Time
	groups     map[string]bool
	rLock      *sync.RWMutex
}

func (u *StoreUser) add(user IUser) {
	u.rLock.Lock()
	defer u.rLock.Unlock()
	u.store[user.GetId()] = user
}
func (u *StoreUser) GetCreateTime() string {
	return u.createTime.Format(util.TimestampFormat)
}

func (u *StoreUser) delete(user IUser) int {
	u.rLock.Lock()
	defer u.rLock.Unlock()
	delete(u.store, user.GetId())
	return len(u.store)
}
func (u *StoreUser) getUsers() []IUser {
	u.rLock.RLock()
	defer u.rLock.RUnlock()
	us := make([]IUser, 0)
	for _, user := range u.store {
		us = append(us, user)
	}
	return us
}
func (u *StoreUser) num() int {
	u.rLock.RLock()
	defer u.rLock.RUnlock()
	return len(u.store)
}
func newUserStore(username string) *StoreUser {
	t := time.Now()
	return &StoreUser{rLock: new(sync.RWMutex), store: make(map[string]IUser), username: username, createTime: &t, groups: make(map[string]bool)}
}

type Store struct {
	uMap  *sync.Map
	num   int32
	rLock *sync.RWMutex
}

func NewStore() *Store {
	return &Store{uMap: new(sync.Map), num: 0, rLock: new(sync.RWMutex)}
}

func (store *Store) AddUser(user IUser) bool {
	username := user.GetUsername()
	store.rLock.Lock()
	v, ok := store.uMap.Load(username)
	if ok {
		us := v.(*StoreUser)
		us.add(user)
	} else {
		us := newUserStore(username)
		us.add(user)
		store.uMap.Store(username, us)
		atomic.AddInt32(&store.num, 1)
		store.rLock.Unlock()
		return true
	}
	store.rLock.Unlock()
	return false
}
func (store *Store) DeleteUser(user IUser) bool {
	username := user.GetUsername()
	v, ok := store.uMap.Load(username)
	if ok {
		us := v.(*StoreUser)
		num := us.delete(user)
		if num == 0 {
			store.rLock.Lock()
			if us.num() == 0 {
				store.uMap.Delete(username)
				atomic.AddInt32(&store.num, -1)
			}
			store.rLock.Unlock()
			return true
		}
	}
	return false
}
func (store *Store) GetUser(username string) ([]IUser, bool) {
	v, ok := store.uMap.Load(username)
	if ok {
		us := v.(*StoreUser)
		return us.getUsers(), true
	}
	return nil, false
}
func (store *Store) UserHasConn() bool {
	return int(store.num) > 0
}
func (store *Store) GetUserNum() int {
	return int(store.num)
}
func (store *Store) Range(f func(username string, user *StoreUser) bool) {
	store.uMap.Range(func(key, value any) bool {
		return f(key.(string), value.(*StoreUser))
	})
}
