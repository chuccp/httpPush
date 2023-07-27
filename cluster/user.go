package cluster

import (
	"github.com/chuccp/httpPush/util"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type cu struct {
	username   string
	machineId  string
	createTime *time.Time
}

func (u *cu) GetId() string {
	return u.username + u.machineId
}
func (u *cu) GetUsername() string {
	return u.username
}
func (u *cu) CreateTime() string {
	return u.createTime.Format(util.TimestampFormat)
}
func newCu(username string, machineId string) *cu {
	u := &cu{username: username, machineId: machineId}
	tu := time.Now()
	u.createTime = &tu
	return u
}

type userStore struct {
	userMap *sync.Map
	num     int32
}

func (us *userStore) AddUser(username string, machineId string) {
	c := newCu(username, machineId)
	_, ok := us.userMap.LoadOrStore(username, c)
	if !ok {
		atomic.AddInt32(&us.num, 1)
	}

	log.Println(us.num)
}
func (us *userStore) DeleteUser(username string) {
	_, ok := us.userMap.LoadAndDelete(username)
	if ok {
		atomic.AddInt32(&us.num, -1)
	}
	log.Println(us.num)
}
func (us *userStore) Num() int32 {
	return us.num
}
func (us *userStore) GetUser(username string) (*cu, bool) {
	v, ok := us.userMap.Load(username)
	if ok {
		return v.(*cu), ok
	}
	return nil, ok
}
func (us *userStore) EachUsers(f func(key string, value *cu) bool) {
	us.userMap.Range(func(key, value interface{}) bool {
		return f(key.(string), value.(*cu))
	})
}

func newUserStore() *userStore {
	return &userStore{userMap: new(sync.Map)}
}
