package cluster

import (
	"encoding/json"
	"github.com/chuccp/httpPush/message"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func MachineId() string {
	f, err := os.OpenFile(".machineId", os.O_RDWR|os.O_CREATE, 0666)
	defer f.Close()
	if err == nil {
		data, err := ioutil.ReadAll(f)
		if err == nil {
			if len(data) == 0 {
				uid := strconv.FormatUint(uint64(message.MsgId()), 36)
				f.Write([]byte(uid))
				return uid
			}
			return strings.TrimSpace(string(data))
		}
	}
	log.Panic("生成机器码错误,请检查程序的读写权限")
	return ""
}

func UnmarshalJsonBody(re *http.Request, v any) error {
	all, err := io.ReadAll(re.Body)
	if err != nil {
		return err
	} else {
		err = json.Unmarshal(all, v)
		if err != nil {
			return err
		}
	}
	return nil
}
