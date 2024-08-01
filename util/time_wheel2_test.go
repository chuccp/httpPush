package util

import (
	"testing"
	"time"
)

func TestTimeWheel2_AfterFunc(t *testing.T) {

	tw2 := NewTimeWheel2(1, 10)
	go tw2.Start()
	tw2.AfterFunc(4, "111111", func(value ...any) {
		println("111")
	})
	time.Sleep(time.Second)
	tw2.AfterFunc(4, "22222", func(value ...any) {
		println("2222")
	})
	time.Sleep(time.Second)
	tw2.AfterFunc(2, "33333", func(value ...any) {
		println("3333")
	})
	time.Sleep(time.Second)
	tw2.AfterFunc(8, "44444", func(value ...any) {
		println("444")
	})
	time.Sleep(time.Second * 60)

}
