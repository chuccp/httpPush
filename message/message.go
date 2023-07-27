package message

import (
	"github.com/chuccp/httpPush/util"
	"math/rand"
	"time"
)

type IMessage interface {
	GetString(byte) string
	GetUint32(byte) uint32
}

type TextMessage struct {
	From  string
	To    string
	Msg   string
	MsgId uint32
}

func (m *TextMessage) GetString(v byte) string {
	if v == Type {
		return "TEXT"
	}
	if v == From {
		return m.From
	}
	if v == To {
		return m.To
	}
	if v == Msg {
		return m.Msg
	}
	return ""
}
func (m *TextMessage) GetUint32(v byte) uint32 {
	if v == MId {
		return m.MsgId
	}
	return 0
}

func NewTextMessage(From string, To string, Msg string) *TextMessage {
	return &TextMessage{From: From, To: To, Msg: Msg, MsgId: MsgId()}
}

func MsgId() uint32 {
	num := rand.Intn(1024)
	return util.Millisecond()<<10 | (uint32(num))
}
func millisecond() uint32 {
	ms := time.Now().UnixNano() / 1e6
	return uint32(ms)
}
func Millisecond() uint32 {
	return millisecond()
}
