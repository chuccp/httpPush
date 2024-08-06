package ex

import (
	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"net/http"
	"sync"
	"time"
)

type User struct {
	user.IUser
	username      string
	remoteAddress string
	liveTime      int
	priority      int
	writer        http.ResponseWriter
	sliceQueue    *util.SliceQueueSafe
	groupIds      []string
	context       *core.Context
	id            string
	onceSend      *OnceSend
	lock          *sync.RWMutex
	lastLiveTime  *time.Time
	last          *time.Time
	expiredTime   *time.Time
	createTime    *time.Time
}
type HttpMessage struct {
	From string
	Body string
}

func newHttpMessage(from string, body string) *HttpMessage {
	return &HttpMessage{From: from, Body: body}
}

func (u *User) GetId() string {
	return u.id
}
func (u *User) RefreshPreExpired() {
	t := time.Now()
	u.last = &t
	tm := t.Add((expiredTime + time.Duration(u.liveTime)*time.Second) * 2)
	u.expiredTime = &tm
}
func (u *User) RefreshExpired() {
	t := time.Now()
	u.last = &t
	tm := t.Add(expiredTime)
	u.expiredTime = &tm
}

func (u *User) waitMessage(tw *util.TimeWheel2) {
	u.lock.Lock()
	send := getOnceSend(u.writer, u.sliceQueue)
	u.onceSend = send
	u.lock.Unlock()
	index := tw.AfterFunc(int32(u.liveTime), u.id, func(value ...any) {
		onceSend, ok := value[0].(*OnceSend)
		if ok {
			onceSend.WriteBlank()
		}
	}, u.onceSend)
	u.onceSend.Wait()
	u.lock.Lock()
	u.onceSend = nil
	freeOnceSend(send)
	u.lock.Unlock()
	tw.DeleteIndexFunc(u.id, index)
}

func (u *User) GetUsername() string {
	return u.username
}
func (u *User) WriteSyncMessage(iMessage message.IMessage) (bool, error) {
	u.lock.Lock()
	if u.onceSend != nil {
		return u.onceSend.WriteAndUnLock(iMessage, func() {
			u.lock.Unlock()
		})
	} else {
		err := u.sliceQueue.Write(iMessage)
		if err != nil {
			u.lock.Unlock()
			return false, err
		}
	}
	u.lock.Unlock()
	return true, nil
}

func (u *User) Close() {}

func (u *User) GetGroupIds() []string {
	return u.groupIds
}
func (u *User) GetRemoteAddress() string {
	return u.remoteAddress
}
func (u *User) SetUsername(username string) {
	u.username = username
}
func (u *User) LastLiveTime() *time.Time {
	return u.lastLiveTime
}
func (u *User) CreateTime() *time.Time {
	return u.createTime
}

func (u *User) GetPriority() int {
	return u.priority
}
func (u *User) GetMachineId() string {
	return ""
}
func (u *User) GetOrderTime() *time.Time {
	return u.lastLiveTime
}
func NewUser(username string, id string, sliceQueue *util.SliceQueueSafe, context *core.Context, writer http.ResponseWriter, re *http.Request) *User {
	u := &User{username: username, id: id, context: context, sliceQueue: sliceQueue, writer: writer, remoteAddress: re.RemoteAddr}
	u.groupIds = util.GetGroupIds(re)
	u.lock = new(sync.RWMutex)
	return u
}
