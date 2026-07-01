package user

import (
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chuccp/httpPush/util"
)

type storeUser struct {
	store      map[string]IUser
	username   string
	createTime *time.Time
	groups     map[string]bool
}

func (u *storeUser) add(user IUser) {
	u.store[user.GetId()] = user
	groupIds := user.GetGroupIds()
	if groupIds != nil {
		for _, s := range groupIds {
			u.groups[s] = true
		}
	}
}
func (u *storeUser) GetCreateTime() string {
	return u.createTime.Format(util.TimestampFormat)
}

func (u *storeUser) delete(user IUser) int {
	delete(u.store, user.GetId())
	return len(u.store)
}
func (u *storeUser) GetUsers() []IUser {
	us := make([]IUser, 0)
	for _, user := range u.store {
		us = append(us, user)
	}
	return us
}
func (u *storeUser) getOrderUser() []IOrderUser {
	us := make([]IOrderUser, 0)
	for _, user := range u.store {
		us = append(us, user)
	}
	sort.Sort(ByAsc(us))
	return us
}
func (u *storeUser) num() int {
	return len(u.store)
}
func (u *storeUser) GetGroupIds() []string {
	v := make([]string, len(u.groups))
	var i = 0
	for k, _ := range u.groups {
		v[i] = k
		i++
	}
	return v
}
func (u *storeUser) IsInGroup(group string) bool {
	_, ok := u.groups[group]
	return ok
}
func newUserStore(username string, newGroupIds []string) *storeUser {
	t := time.Now()
	groups := make(map[string]bool)
	for _, id := range newGroupIds {
		groups[id] = true
	}
	return &storeUser{
		store:      make(map[string]IUser),
		username:   username,
		createTime: &t,
		groups:     groups,
	}
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
	store.rLock.Lock()
	defer store.rLock.Unlock()
	username := user.GetUsername()
	v, ok := store.uMap.Load(username)
	newGroupIds := user.GetGroupIds()
	if ok {
		us := v.(*storeUser)
		oldGroupIds := us.GetGroupIds()
		newGroupIdsMap := make(map[string]bool)
		for _, newGroupId := range newGroupIds {
			newGroupIdsMap[newGroupId] = true
		}
		for _, groupId := range oldGroupIds {
			if _, ok := newGroupIdsMap[groupId]; !ok {
				gp, ok := store.gMap.Load(groupId)
				if ok {
					us := gp.(*StoreGroup)
					us.RemoteUser(user)
				}
			}
		}
		us.add(user)
	} else {
		us := newUserStore(username, newGroupIds)
		us.add(user)
		store.uMap.Store(username, us)
		atomic.AddInt32(&store.num, 1)
		for _, groupId := range newGroupIds {
			gp, ok := store.gMap.Load(groupId)
			if ok {
				storeGroup := gp.(*StoreGroup)
				storeGroup.AddUser(user)
			} else {
				storeGroup := NewStoreGroup()
				storeGroup.AddUser(user)
				store.gMap.Store(groupId, storeGroup)
			}
		}

		return true
	}
	return false
}
func (store *Store) DeleteUser(user IUser) bool {
	store.rLock.Lock()
	defer store.rLock.Unlock()
	username := user.GetUsername()
	v, ok := store.uMap.Load(username)
	if ok {
		us := v.(*storeUser)
		us.delete(user)
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
		return true

	}
	return false
}
func (store *Store) GetUser(username string) ([]IUser, bool) {
	v, ok := store.uMap.Load(username)
	if ok {
		store.rLock.RLock()
		defer store.rLock.RUnlock()
		us := v.(*storeUser)
		return us.GetUsers(), true
	}
	return nil, false
}

func (store *Store) GetUserCreateTime(username string) *time.Time {
	v, ok := store.uMap.Load(username)
	if ok {
		store.rLock.RLock()
		defer store.rLock.RUnlock()
		us := v.(*storeUser)
		return us.createTime
	}
	return nil
}

func (store *Store) HasLocalUser(username string) bool {
	v, ok := store.uMap.Load(username)
	if ok {
		store.rLock.RLock()
		defer store.rLock.RUnlock()
		us := v.(*storeUser)
		return us.num() > 0
	}
	return false
}

func (store *Store) GetOrderUser(username string) []IOrderUser {

	v, ok := store.uMap.Load(username)
	if ok {
		store.rLock.RLock()
		defer store.rLock.RUnlock()
		us := v.(*storeUser)
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
	store.rLock.RLock()
	defer store.rLock.RUnlock()
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

type StoreUserInfo struct {
	UserName   string
	CreateTime string
	Users      []IUser
}

func (store *Store) Range(f func(username string, user *StoreUserInfo) bool) {
	store.uMap.Range(func(key, value any) bool {
		store.rLock.RLock()
		us := value.(*storeUser)
		users := us.GetUsers()
		username := key.(string)
		pu := &StoreUserInfo{
			UserName:   username,
			CreateTime: us.GetCreateTime(),
			Users:      users,
		}
		store.rLock.RUnlock()
		return f(username, pu)
	})
}
