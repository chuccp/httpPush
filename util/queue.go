package util

import (
	"sync"
)

type Queue struct {
	sliceQueue *SliceQueue
	lock       *sync.RWMutex
	waitNum    int32
	flag       chan bool
}

func (queue *Queue) Offer(value any) error {
	queue.lock.Lock()
	err := queue.sliceQueue.Write(value)
	if queue.waitNum > 0 {
		queue.waitNum--
		queue.lock.Unlock()
		queue.flag <- true
	} else {
		queue.lock.Unlock()
	}
	return err
}
func (queue *Queue) DequeueTimer(timer *Timer) (value any, hasValue bool) {
	for {
		queue.lock.Lock()
		v, err := queue.sliceQueue.Read()
		if err == nil {
			queue.lock.Unlock()
			timer.Close()
			return v, true
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			select {
			case fa := <-queue.flag:
				{
					if fa {
						continue
					} else {
						timer.Close()
						return nil, false
					}
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
func (queue *Queue) Dequeue() (value any, hasValue bool) {
	for {
		queue.lock.Lock()
		v, err := queue.sliceQueue.Read()
		if err == nil {
			queue.lock.Unlock()
			if v == nil {
				return nil, false
			}
			return v, true
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			select {
			case fa := <-queue.flag:
				{
					if fa {
						continue
					} else {
						return nil, false
					}
				}
			}
		}
	}
}

var poolQueue = &sync.Pool{
	New: func() interface{} {
		return &Queue{lock: new(sync.RWMutex), flag: make(chan bool), waitNum: 0}
	},
}

func GetQueue() *Queue {
	queue := poolQueue.Get().(*Queue)
	queue.sliceQueue = GetSliceQueue()
	if queue.waitNum > 0 {
		close(queue.flag)
		queue.flag = make(chan bool)
		queue.waitNum = 0
	}
	return queue
}
func FreeQueue(queue *Queue) {
	FreeSliceQueue(queue.sliceQueue)
	poolQueue.Put(queue)
}
