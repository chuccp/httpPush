package message

import (
	"github.com/chuccp/httpPush/util"
	"math/rand"
	"time"
)

type IMessage interface {
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
