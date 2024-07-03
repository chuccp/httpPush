package util

import (
	"log"
	"testing"
	"time"
)

func TestAAAAAAA(t *testing.T) {
	tw := NewTimeWheel(3, 10)

	for {
		timer := tw.NewTimer(4)
		go func() {
			<-timer.C
		}()
		time.Sleep(time.Second * 1)
	}

	log.Println("end", time.Now())
	time.Sleep(time.Hour * 1)
}
