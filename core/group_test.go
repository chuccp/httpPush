package core

import (
	"github.com/panjf2000/ants/v2"
	"testing"
	"time"
)

func xxx(num int, f func()) {

}

func TestWaitGroup(t *testing.T) {

	ants, _ := ants.NewPool(5)

	for i := 0; i < 100; i++ {
		ants.Submit(func() {
			time.Sleep(time.Second)
			println("==============")
		})
	}
}
