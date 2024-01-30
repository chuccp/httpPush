package core

import (
	"sync"
	"testing"
)

func xxx(num int, f func()) {
	func() {
		println("==========", num)
		//time.Sleep(time.Second)
		f()
	}()
}

func TestWaitGroup(t *testing.T) {

	numbers := []int{1, 2, 3, 4, 5}

	waitGroup := new(sync.WaitGroup)

	go func() {

		for _, number := range numbers {
			waitGroup.Add(1)
			xxx(number, func() {

				waitGroup.Done()
				waitGroup.Done()
				waitGroup.Done()

			})
		}
	}()

	t.Log("======Wait=================")
	waitGroup.Wait()

	t.Log("=======================")

}
