package cluster

import (
	"github.com/chuccp/httpPush/message"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func MachineId() string {
	f, err := os.OpenFile(".machineId", os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Panic("生成机器码错误,请检查程序的读写权限")
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		log.Panic("生成机器码错误,请检查程序的读写权限")
	}
	if len(data) == 0 {
		uid := strconv.FormatUint(uint64(message.MsgId()), 36)
		f.Write([]byte(uid))
		return uid
	}
	return strings.TrimSpace(string(data))
}
