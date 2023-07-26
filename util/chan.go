package util

import (
	"sync"
)

var poolBoolChan = &sync.Pool{
	New: func() interface{} {
		return make(chan bool)
	},
}
func GetChanBool() chan bool {
	flag := poolBoolChan.Get().(chan bool)
	return flag
}
func FreeChanBool(flag chan bool) {
	poolBoolChan.Put(flag)
}