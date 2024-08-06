package ex

import (
	"encoding/json"
	"github.com/chuccp/httpPush/message"
	"github.com/chuccp/httpPush/util"
	"io"
	"sync"
)

func (s *OnceSend) messageToBytes(iMessage message.IMessage) []byte {
	ht := newHttpMessage(
		iMessage.GetString(message.From),
		iMessage.GetString(message.Msg))
	hts := []*HttpMessage{ht}
	data, _ := json.Marshal(hts)
	return data
}

type OnceSend struct {
	writer     io.Writer
	sliceQueue *util.SliceQueueSafe
	fa         chan bool
	isWait     bool
	isWrite    bool
	lock       *sync.RWMutex
}

func (s *OnceSend) write(iMessage message.IMessage) (n bool, err error) {
	if !s.isWrite {
		s.isWrite = true
		if iMessage != nil {
			bytes := s.messageToBytes(iMessage)
			_, err := s.writer.Write(bytes)
			return err == nil, err
		} else {
			_, err := s.writer.Write([]byte("[]"))
			return err == nil, err
		}
	} else {
		if iMessage != nil {
			s.sliceQueue.Write(iMessage)
		}
	}
	return true, nil
}

func (s *OnceSend) Wait() {
	s.lock.Lock()
	read, err := s.sliceQueue.Read()
	if err != nil {
		s.isWait = true
		s.lock.Unlock()
		<-s.fa
	} else {
		v, ok := read.(message.IMessage)
		if ok {
			s.write(v)
		}
		s.lock.Unlock()
	}
}

func (s *OnceSend) WriteBlank(f func()) {
	s.lock.Lock()
	if s.isWait {
		s.isWait = false
		s.write(nil)
		s.lock.Unlock()
		f()
		s.fa <- true
	} else {
		s.lock.Unlock()
		f()
	}
}

func (s *OnceSend) WriteAndUnLock(iMessage message.IMessage, f func()) (n bool, err error) {
	s.lock.Lock()
	if s.isWait {
		s.isWait = false
		_, err := s.write(iMessage)
		s.lock.Unlock()
		f()
		s.fa <- true
		return err == nil, err
	} else {
		err := s.sliceQueue.Write(iMessage)
		s.lock.Unlock()
		f()
		return err == nil, err
	}
}

var poolOnceSend = &sync.Pool{
	New: func() interface{} {
		return &OnceSend{fa: make(chan bool), lock: new(sync.RWMutex)}
	},
}

func getOnceSend(writer io.Writer, sliceQueue *util.SliceQueueSafe) *OnceSend {
	onceSend := poolOnceSend.Get().(*OnceSend)
	onceSend.writer = writer
	onceSend.sliceQueue = sliceQueue
	onceSend.isWrite = false
	onceSend.isWait = false
	return onceSend
}
func freeOnceSend(onceSend *OnceSend) {
	poolOnceSend.Put(onceSend)
}
