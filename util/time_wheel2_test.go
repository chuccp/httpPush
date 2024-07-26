package util

import (
	"strconv"
	"testing"
	"time"
)

func TestTimeWheel2_AfterFunc(t *testing.T) {

	tw2 := NewTimeWheel2(1, 60)
	go tw2.Start()
	var index = 0
	for {
		index++
		tw2.AfterFunc(4, strconv.Itoa(index), func() {
			t.Log(strconv.Itoa(index))
		})
		index++
		tw2.AfterFunc(4, strconv.Itoa(index), func() {
			t.Log(strconv.Itoa(index))
		})
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second * 180)
}
