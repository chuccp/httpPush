package util

import (
	"testing"
	"time"
)

func TestName222(t *testing.T) {

	tk := time.NewTicker(time.Second * 1)
	defer tk.Stop()
	index := 0
	for {
		select {
		case <-tk.C:
			index++
			t.Log("======", time.Now())
			if index%4 == 0 {
				time.Sleep(time.Second * 10)
			} else {
				time.Sleep(time.Millisecond * 1)
			}

		}

	}
}
