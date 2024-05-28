package util

import (
	"container/list"
	"context"
	"errors"
	"sync"
)

type Queue struct {
	list    *list.List
	lock    *sync.RWMutex
	waitNum int32
	flag    chan bool
}

func NewQueue() *Queue {
	return &Queue{list: list.New(), lock: new(sync.RWMutex), flag: make(chan bool)}
}

func (queue *Queue) Offer(value interface{}) (nu int32) {
	queue.lock.Lock()
	queue.list.PushBack(value)
	num := queue.list.Len()
	if queue.waitNum > 0 {
		queue.waitNum--
		queue.lock.Unlock()
		queue.flag <- true
	} else {
		queue.lock.Unlock()
	}
	return int32(num)
}
func (queue *Queue) Dequeue(ctx context.Context) (value interface{}, num int32, hasClose bool) {

	for {
		queue.lock.Lock()
		num := queue.list.Len()
		if num > 0 {
			ele := queue.list.Front()
			queue.list.Remove(ele)
			queue.lock.Unlock()
			return ele.Value, int32(num - 1), false
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			go func() {
				<-ctx.Done()
				queue.lock.Lock()
				err := ctx.Err()
				if errors.Is(err, context.DeadlineExceeded) {
					if queue.waitNum > 0 {
						queue.waitNum--
						queue.lock.Unlock()
						queue.flag <- false
					} else {
						queue.lock.Unlock()
					}
				} else {
					queue.lock.Unlock()
				}
			}()
			fa := <-queue.flag
			if !fa {
				return nil, 0, true
			}
		}
	}

}
func (queue *Queue) Poll() (value interface{}, num int32) {
	for {
		queue.lock.Lock()
		num := queue.list.Len()
		if num > 0 {
			ele := queue.list.Front()
			queue.list.Remove(ele)
			queue.lock.Unlock()
			return ele.Value, int32(num - 1)
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			<-queue.flag
		}
	}
}
