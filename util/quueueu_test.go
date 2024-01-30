package util

import (
	"testing"
	"time"
)

func num(num int, queue *vQueue) {
	queue.Enqueue(num)
	time.Sleep(time.Second)
}
func TestAAAA(t *testing.T) {

	queu := NewVQueue2()

	go func() {

		for i := 0; i < 1000; i++ {
			time.Sleep(time.Second)
			go num(i, queu)
		}

		//queu.Enqueue(1)
		////time.Sleep(time.Second)
		//queu.Enqueue(2)
		////time.Sleep(time.Second)
		//queu.Enqueue(3)
		//queu.Enqueue(1)
		////time.Sleep(time.Second)
		//queu.Enqueue(2)
		////time.Sleep(time.Second)
		//queu.Enqueue(3)
		//queu.Enqueue(1)
		////time.Sleep(time.Second)
		//queu.Enqueue(2)
		////time.Sleep(time.Second)
		//queu.Enqueue(3)
		//time.Sleep(time.Second * 10)
	}()

	go func() {
		for {
			t.Log(queu.Dequeue())
		}
	}()

	// 等待goroutine执行完成
	time.Sleep(time.Second * 10)
}
