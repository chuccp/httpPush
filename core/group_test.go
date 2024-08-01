package core

import (
	"sync"
	"testing"
	"time"
)

func xxx(num int, f func()) {

}

func TestWaitGroup(t *testing.T) {
	wg := new(sync.WaitGroup)
	once := new(sync.Once)
	wg.Add(1)
	go func() {
		once.Do(func() {
			wg.Done()
		})
	}()
	go func() {
		once.Do(func() {
			wg.Done()
		})
	}()
	wg.Wait()
	time.Sleep(time.Second)
}
