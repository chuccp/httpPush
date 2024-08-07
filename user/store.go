package user

import (
	"github.com/chuccp/httpPush/util"
	"sort"
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
	groupIds := user.GetGroupIds()
	if groupIds != nil {
		for _, s := range groupIds {
			u.groups[s] = true
		}
	}
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
func (u *StoreUser) getOrderUser() []IOrderUser {
	u.rLock.RLock()
	defer u.rLock.RUnlock()
	us := make([]IOrderUser, 0)
	for _, user := range u.store {
		us = append(us, user)
	}
	sort.Sort(ByAsc(us))
	return us
}
func (u *StoreUser) num() int {
	u.rLock.RLock()
	defer u.rLock.RUnlock()
	return len(u.store)
}
func (u *StoreUser) GetGroupIds() []string {
	u.rLock.RLock()
	defer u.rLock.RUnlock()
	v := make([]string, len(u.groups))
	var i = 0
	for k, _ := range u.groups {
		v[i] = k
		i++
	}
	return v
}
func (u *StoreUser) IsInGroup(group string) bool {
	u.rLock.RLock()
	defer u.rLock.RUnlock()
	_, ok := u.groups[group]
	return ok
}
func newUserStore(username string) *StoreUser {
	t := time.Now()
	return &StoreUser{rLock: new(sync.RWMutex), store: make(map[string]IUser), username: username, createTime: &t, groups: make(map[string]bool)}
}

type StoreGroup struct {
	uMap *sync.Map
	num  int
}

func NewStoreGroup() *StoreGroup {
	return &StoreGroup{uMap: new(sync.Map)}
}

func (storeGroup *StoreGroup) AddUser(user IUser) {
	v, ok := storeGroup.uMap.Load(user.GetUsername())
	if ok {
		group := v.(*Group)
		group.lastLiveTime = user.LastLiveTime()
	} else {
		storeGroup.num++
		group := NewGroup(user)
		storeGroup.uMap.Store(user.GetUsername(), group)
	}
}
func (storeGroup *StoreGroup) RangeUser(f func(string) bool) {
	storeGroup.uMap.Range(func(key, value any) bool {
		return f(key.(string))
	})
}
func (storeGroup *StoreGroup) RemoteUser(user IUser) {
	storeGroup.num--
	storeGroup.uMap.Delete(user.GetUsername())
}
func (storeGroup *StoreGroup) GetNum() int {
	return storeGroup.num
}

type Store struct {
	uMap  *sync.Map
	gMap  *sync.Map
	num   int32
	rLock *sync.RWMutex
}

func NewStore() *Store {
	return &Store{gMap: new(sync.Map), uMap: new(sync.Map), num: 0, rLock: new(sync.RWMutex)}
}

func (store *Store) AddUser(user IUser) bool {
	username := user.GetUsername()
	store.rLock.Lock()
	v, ok := store.uMap.Load(username)
	groupIds := user.GetGroupIds()
	if groupIds != nil {
		for _, groupId := range groupIds {
			if len(groupId) > 0 {
				gp, ok := store.gMap.Load(groupId)
				if ok {
					storeGroup := gp.(*StoreGroup)
					storeGroup.AddUser(user)
				} else {
					storeGroup := NewStoreGroup()
					store.gMap.Store(groupId, storeGroup)
					storeGroup.AddUser(user)
				}
			}
		}
	}
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
				groupIds := us.GetGroupIds()
				if groupIds != nil {
					for _, groupId := range groupIds {
						gp, ok := store.gMap.Load(groupId)
						if ok {
							storeGroup := gp.(*StoreGroup)
							storeGroup.RemoteUser(user)
						}
					}
				}
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

func (store *Store) GetUserCreateTime(username string) *time.Time {
	v, ok := store.uMap.Load(username)
	if ok {
		us := v.(*StoreUser)
		return us.createTime
	}
	return nil
}

func (store *Store) HasLocalUser(username string) bool {
	v, ok := store.uMap.Load(username)
	if ok {
		us := v.(*StoreUser)
		return us.num() > 0
	}
	return false
}

func (store *Store) GetOrderUser(username string) []IOrderUser {
	v, ok := store.uMap.Load(username)
	if ok {
		us := v.(*StoreUser)
		return us.getOrderUser()
	}
	return make([]IOrderUser, 0)
}

func (store *Store) RangeGroupUser(groupId string, f func(username string) bool) {
	gp, ok := store.gMap.Load(groupId)
	if ok {
		storeGroup := gp.(*StoreGroup)
		storeGroup.RangeUser(f)
	}
}
func (store *Store) RangeAllUser(f func(username string) bool) {
	store.uMap.Range(func(key, _ any) bool {
		v, ok := key.(string)
		if ok {
			f(v)
		}
		return true
	})
}

func (store *Store) QueryGroupsUser(groupIds ...string) *GroupUser {
	groupUser := NewGroupUser()
	for _, groupId := range groupIds {
		gp, ok := store.gMap.Load(groupId)
		if ok {
			storeGroup := gp.(*StoreGroup)
			storeGroup.RangeUser(func(s string) bool {
				groupUser.AddUser(s)
				return true
			})
		}
	}
	return groupUser
}

func (store *Store) UserHasConn() bool {
	return int(store.num) > 0
}

func (store *Store) AllGroupInfo() map[string]int {
	groupMap := make(map[string]int)
	store.gMap.Range(func(key, value any) bool {
		sg, ok1 := value.(*StoreGroup)
		k, ok2 := key.(string)
		if ok1 && ok2 {
			groupMap[k] = sg.num
		}
		return true
	})
	return groupMap
}
func (store *Store) GetUserNum() int {
	return int(store.num)
}
func (store *Store) Range(f func(username string, user *StoreUser) bool) {
	store.uMap.Range(func(key, value any) bool {
		return f(key.(string), value.(*StoreUser))
	})
}
