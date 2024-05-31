package util

import (
	"os"
	"os/signal"
	"syscall"
	"testing"
)

func Test2Queue(t *testing.T) {

	queue := NewQueue()

	go func() {

		queue.Offer(1)
		queue.Offer(2)
		queue.Offer(3)
		queue.Offer(4)
		queue.Offer(5)

	}()

	go func() {
		for {
			v := queue.Poll()
			t.Log(v)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGBUS)
	<-sig
}
