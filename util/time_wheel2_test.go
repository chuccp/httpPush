package util

import (
	"log"
	"testing"
	"time"
)

func TestTimeWheel2_AfterFunc(t *testing.T) {

	chanBool := make(chan bool)

	go func() {
		for {
			fa := <-chanBool
			log.Println("00000000000", fa)

		}
	}()
	time.Sleep(time.Second)
	go func() {
		for {
			fa := <-chanBool
			log.Println("11111111", fa)

		}
	}()
	time.Sleep(time.Second)
	for {
		chanBool <- true
		time.Sleep(time.Second * 2)
	}
	time.Sleep(10 * time.Second)
}
