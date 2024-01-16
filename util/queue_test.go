package util

import (
	"testing"
	"time"
)

func Test2Queue(t *testing.T) {
	sendQueue := NewQueue()

	go func() {
		for {
			v, num := sendQueue.Poll()
			t.Log(v, num)
		}
	}()

	sendQueue.Offer("111")
	sendQueue.Offer("111")

	time.Sleep(time.Second * 5)
}
