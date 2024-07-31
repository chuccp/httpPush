package util

import (
	"container/list"
	"sync"
)

type Queue2 struct {
	list    *list.List
	lock    *sync.RWMutex
	waitNum int32
	flag    chan bool
}

func NewQueue() *Queue2 {
	return &Queue2{list: list.New(), lock: new(sync.RWMutex), flag: make(chan bool)}
}

func (queue *Queue2) Offer(value interface{}) {
	queue.lock.Lock()
	queue.list.PushFront(value)
	if queue.waitNum > 0 {
		queue.waitNum--
		queue.lock.Unlock()
		queue.flag <- true
	} else {
		queue.lock.Unlock()
	}
}
func (queue *Queue2) DequeueTimer(timer *Timer) (value interface{}, hasValue bool) {
	for {
		queue.lock.Lock()
		ele := queue.list.Back()
		if ele != nil {
			queue.list.Remove(ele)
			queue.lock.Unlock()
			timer.Close()
			return ele.Value, true
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			select {
			case <-queue.flag:
				{
					continue
				}
			case <-timer.C:
				{
					queue.lock.Lock()
					if queue.waitNum > 0 {
						queue.waitNum--
						queue.lock.Unlock()
					} else {
						queue.lock.Unlock()
					}
					timer.Close()
					return nil, false
				}
			}
		}
	}
}
