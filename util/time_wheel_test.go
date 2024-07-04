package util

import (
	"log"
	"testing"
	"time"
)

func TestAAAAAAA(t *testing.T) {

	for {
		fa := make(chan bool)
		go func() {
			time.Sleep(time.Second * 1)
			ccc := <-fa
			log.Println(ccc)
		}()
		close(fa)
		println("11111111111")
		time.Sleep(time.Second * 10)
	}

	log.Println("end", time.Now())
	time.Sleep(time.Hour * 1)
}
