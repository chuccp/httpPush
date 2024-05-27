package cluster

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"go.uber.org/zap"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const expiredTime = 10 * time.Minute

var poolCuStore = &sync.Pool{
	New: func() interface{} {
		return &cuStore{}
	},
}

func getNewCuStore(username string) *cuStore {
	flag := poolCuStore.Get().(*cuStore)
	flag.username = username
	t := time.Now()
	flag.createTime = &t
	flag.store = make(map[string]*clientUser)
	return flag
}
func freeCuStore(cuStore *cuStore) {
	poolCuStore.Put(cuStore)
}

type cuStore struct {
	store      map[string]*clientUser
	username   string
	createTime *time.Time
}

func (cs *cuStore) addUser(username string, machineId string, clientOperate *ClientOperate) {
	cu := newCu(username, machineId, clientOperate)
	cs.store[machineId] = cu
}
func (cs *cuStore) renewUser(username string, machineId string, clientOperate *ClientOperate) {
	v, ok := cs.store[machineId]
	if !ok {
		cu := newCu(username, machineId, clientOperate)
		cs.store[machineId] = cu
	} else {
		t := time.Now()
		v.lastLiveTime = &t
		ext := t.Add(expiredTime)
		v.expiredTime = &ext
	}
}
func (cs *cuStore) getUser() []user.IOrderUser {
	us := make([]user.IOrderUser, 0)
	if len(cs.store) > 0 {
		for _, user := range cs.store {
			us = append(us, user)
		}
		sort.Sort(user.ByAsc(us))
	}
	return us

}

type clientUser struct {
	username      string
	machineId     string
	priority      int
	createTime    *time.Time
	lastLiveTime  *time.Time
	expiredTime   *time.Time
	clientOperate *ClientOperate
}

func (u *clientUser) isExpiredTime(now *time.Time) bool {
	return u.expiredTime != nil && u.expiredTime.Before(*now)
}

func (u *clientUser) GetOrderTime() *time.Time {
	if u.lastLiveTime != nil {
		return u.lastLiveTime
	}
	return u.createTime
}

func (u *clientUser) GetMachineId() string {
	return u.machineId
}

func (u *clientUser) GetId() string {
	return u.username + u.machineId
}

func (u *clientUser) GetPriority() int {
	return u.priority
}
func (u *clientUser) WriteMessage(msg message.IMessage, writeFunc user.WriteCallBackFunc) {
	switch t := msg.(type) {
	case *message.TextMessage:
		{
			cl, ok := u.clientOperate.getClient(u.machineId)
			if ok {
				err := cl.sendTextMsg(t)
				if err == nil {
					u.priority = 0
					writeFunc(nil, true)
					return
				}
			}
		}
	}
	if u.priority < 5 {
		u.priority = u.priority + 1
	}
	writeFunc(nil, false)
}

func (u *clientUser) GetUsername() string {
	return u.username
}
func (u *clientUser) CreateTime() string {
	return u.createTime.Format(util.TimestampFormat)
}
func newCu(username string, machineId string, clientOperate *ClientOperate) *clientUser {
	u := &clientUser{username: username, machineId: machineId, clientOperate: clientOperate, priority: 0}
	tu := time.Now()
	u.createTime = &tu
	u.lastLiveTime = &tu
	ext := tu.Add(expiredTime)
	u.expiredTime = &ext
	return u
}

type userStore struct {
	userMap *sync.Map
	num     int32
	rLock   *sync.RWMutex
	context *core.Context
}

func (us *userStore) AddUser(username string, machineId string, clientOperate *ClientOperate) {
	us.rLock.Lock()
	defer us.rLock.Unlock()
	cus, ok := us.userMap.Load(username)
	if !ok {
		us.context.GetLog().Debug("AddUser", zap.String("username", username), zap.Int("us.num", int(us.num)))
		atomic.AddInt32(&us.num, 1)
		cus := getNewCuStore(username)
		cus.addUser(username, machineId, clientOperate)
		us.userMap.Store(username, cus)
	} else {
		sc := cus.(*cuStore)
		sc.addUser(username, machineId, clientOperate)
	}
}

func (us *userStore) RefreshUser(username string, machineId string, clientOperate *ClientOperate) {
	us.rLock.Lock()
	defer us.rLock.Unlock()
	cus, ok := us.userMap.Load(username)
	if !ok {
		atomic.AddInt32(&us.num, 1)
		cus := getNewCuStore(username)
		cus.addUser(username, machineId, clientOperate)
		us.userMap.Store(username, cus)
	} else {
		sc := cus.(*cuStore)
		sc.renewUser(username, machineId, clientOperate)
	}
}

func (us *userStore) DeleteUser(username string, machineIds []string) {
	if len(machineIds) > 0 {
		us.rLock.Lock()
		defer us.rLock.Unlock()
		cus, ok := us.userMap.Load(username)
		if ok {
			sc := cus.(*cuStore)
			for _, machineId := range machineIds {
				delete(sc.store, machineId)
			}
			if len(sc.store) == 0 {
				us.userMap.Delete(username)
				freeCuStore(sc)
				atomic.AddInt32(&us.num, -1)
			}
		}
	}
}

func (us *userStore) DeleteExpiredUser(username string, now *time.Time) {
	us.rLock.Lock()
	defer us.rLock.Unlock()
	cu, ok := us.userMap.Load(username)
	if ok {
		sc := cu.(*cuStore)
		for machineId, c := range sc.store {
			if c.isExpiredTime(now) {
				delete(sc.store, machineId)
				if len(sc.store) == 0 {
					us.userMap.Delete(username)
					freeCuStore(sc)
					atomic.AddInt32(&us.num, -1)
				}
			}
		}
	}
}

func (us *userStore) Num() int32 {
	return us.num
}

func (us *userStore) ClearTimeOutUser(now time.Time) {
	us.userMap.Range(func(key, value interface{}) bool {
		cu := value.(*cuStore)
		us.DeleteExpiredUser(cu.username, &now)
		return true
	})
}

func (us *userStore) GetOrderUser(username string) []user.IOrderUser {
	us.rLock.RLock()
	defer us.rLock.RUnlock()
	cvs, ok := us.userMap.Load(username)
	if ok {
		sc := cvs.(*cuStore)
		return sc.getUser()
	}
	u := make([]user.IOrderUser, 0)
	return u
}
func newUserStore(context *core.Context) *userStore {
	return &userStore{userMap: new(sync.Map), rLock: new(sync.RWMutex), context: context}
}
