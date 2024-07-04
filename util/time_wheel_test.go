package util

import (
	"log"
	"testing"
	"time"
)

func TestAAAAAAA(t *testing.T) {

	for {
		fa := make(chan bool)
		close(fa)
		fa <- true

		time.Sleep(time.Second * 1)
	}

	log.Println("end", time.Now())
	time.Sleep(time.Hour * 1)
}
