package ws

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/chuccp/httpPush/core"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/user"
	"github.com/chuccp/httpPush/util"
	"github.com/gorilla/websocket"
)

type User struct {
	user.IUser
	username      string
	remoteAddress string
	groupIds      []string
	context       *core.App
	id            string
	conn          *websocket.Conn
	writeCh       chan []byte
	lastLiveTime  *time.Time
	createTime    *time.Time
	priority      int
}

type wsMessage struct {
	From string `json:"from"`
	Body string `json:"body"`
}

func (u *User) GetId() string            { return u.id }
func (u *User) GetUsername() string      { return u.username }
func (u *User) GetRemoteAddress() string { return u.remoteAddress }
func (u *User) GetGroupIds() []string    { return u.groupIds }
func (u *User) SetUsername(name string)  { u.username = name }
func (u *User) LastLiveTime() *time.Time { return u.lastLiveTime }
func (u *User) CreateTime() *time.Time   { return u.createTime }
func (u *User) GetPriority() int         { return u.priority }
func (u *User) GetMachineId() string     { return "" }
func (u *User) GetOrderTime() *time.Time { return u.lastLiveTime }

func (u *User) Close() {
	close(u.writeCh)
}

// WriteSyncMessage 推送消息到 WebSocket（通过 write pump）
func (u *User) WriteSyncMessage(iMessage message.IMessage) (bool, error) {
	t := time.Now()
	u.lastLiveTime = &t

	msg := &wsMessage{
		From: iMessage.GetString(message.From),
		Body: iMessage.GetString(message.Msg),
	}
	data, err := json.Marshal([]*wsMessage{msg})
	if err != nil {
		return false, err
	}

	select {
	case u.writeCh <- data:
		return true, nil
	default:
		return false, nil
	}
}

func NewUser(username string, id string, context *core.App, conn *websocket.Conn, writeCh chan []byte, re *http.Request) *User {
	t := time.Now()
	u := &User{
		username:      username,
		id:            id,
		context:       context,
		conn:          conn,
		writeCh:       writeCh,
		remoteAddress: re.RemoteAddr,
		lastLiveTime:  &t,
		createTime:    &t,
	}
	u.groupIds = util.GetGroupIds(re)
	return u
}
