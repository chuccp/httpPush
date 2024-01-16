package cluster

import (
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"sync"
	"sync/atomic"
	"time"
)

type cuStore struct {
	store      map[string]*cu
	username   string
	createTime *time.Time
}

func newCuStore(username string) *cuStore {
	t := time.Now()
	return &cuStore{store: make(map[string]*cu), username: username, createTime: &t}
}
func (cs *cuStore) addUser(username string, machineId string, clientOperate *ClientOperate) {
	cu := newCu(username, machineId, clientOperate)
	cs.store[machineId] = cu
}
func (cs *cuStore) getUser() []user.IOrderUser {
	us := make([]user.IOrderUser, 0)
	for _, user := range cs.store {
		us = append(us, user)
	}
	return us

}
func (cs *cuStore) deleteUser(machineId string) {
	delete(cs.store, machineId)
}
func (cs *cuStore) num() int {
	return len(cs.store)
}

type cu struct {
	username      string
	machineId     string
	priority      int
	createTime    *time.Time
	lastLiveTime  *time.Time
	clientOperate *ClientOperate
}

func (u *cu) GetOrderTime() *time.Time {
	if u.lastLiveTime != nil {
		return u.lastLiveTime
	}
	return u.createTime
}

func (u *cu) GetMachineId() string {
	return u.machineId
}

func (u *cu) GetId() string {
	return u.username + u.machineId
}

func (u *cu) GetPriority() int {
	return u.priority
}
func (u *cu) WriteMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	switch t := msg.(type) {
	case *message.TextMessage:
		{
			cl, ok := u.clientOperate.getClient(u.machineId)
			if ok {
				err := cl.sendTextMsg(t)
				if err == nil {
					t := time.Now()
					u.lastLiveTime = &t
					u.priority = 0
					writeFunc(nil, true)
					return
				}
			}
		}
	}
	u.priority = 1
	writeFunc(nil, false)
}

func (u *cu) GetUsername() string {
	return u.username
}
func (u *cu) CreateTime() string {
	return u.createTime.Format(util.TimestampFormat)
}
func newCu(username string, machineId string, clientOperate *ClientOperate) *cu {
	u := &cu{username: username, machineId: machineId, clientOperate: clientOperate}
	tu := time.Now()
	u.createTime = &tu
	return u
}

type userStore struct {
	userMap *sync.Map
	num     int32
	rLock   *sync.RWMutex
}

func (us *userStore) AddUser(username string, machineId string, clientOperate *ClientOperate) {
	us.rLock.Lock()
	defer us.rLock.Unlock()
	cus, ok := us.userMap.Load(username)
	if !ok {
		atomic.AddInt32(&us.num, 1)
		cus := newCuStore(username)
		cus.addUser(username, machineId, clientOperate)
		us.userMap.Store(username, cus)
	} else {
		sc := cus.(*cuStore)
		sc.addUser(username, machineId, clientOperate)
	}

}
func (us *userStore) DeleteUser(username string, machineId string) {
	us.rLock.Lock()
	defer us.rLock.Unlock()
	cu, ok := us.userMap.LoadAndDelete(username)
	if ok {
		sc := cu.(*cuStore)
		sc.deleteUser(machineId)
		if sc.num() == 0 {
			atomic.AddInt32(&us.num, -1)
		}
	}
}
func (us *userStore) Num() int32 {
	return us.num
}
func (us *userStore) GetOrderUser(username string) ([]user.IOrderUser, bool) {
	us.rLock.RLock()
	defer us.rLock.RUnlock()
	cvs, ok := us.userMap.Load(username)
	if ok {
		sc := cvs.(*cuStore)
		return sc.getUser(), true
	}
	return nil, ok
}
func (us *userStore) EachUsers(f func(key string, value *cu) bool) {
	us.userMap.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*cu))
	})
}

func newUserStore() *userStore {
	return &userStore{userMap: new(sync.Map), rLock: new(sync.RWMutex)}
}
