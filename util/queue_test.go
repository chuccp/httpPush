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
		queue.Offer(1)
		queue.Offer(1)
		queue.Offer(1)
		queue.Offer(1)

	}()

	go func() {
		for {
			v, num := queue.Poll()
			t.Log(v, num)
		}
	}()
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGBUS)
	<-sig
}
