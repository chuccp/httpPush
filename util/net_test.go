package util

import (
	"context"
	"log"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	queue := NewQueue()

	for {
		ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
		_, m, flag := queue.Dequeue(ctx)
		if flag {
			log.Println(m)
		} else {
			cancelFunc()
		}
	}
}
