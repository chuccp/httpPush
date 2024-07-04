package util

import (
	"github.com/panjf2000/ants/v2"
	"sync"
)

type Queue struct {
	sliceQueue *SliceQueue
	lock       *sync.RWMutex
	waitNum    int32
	flag       chan bool
}

func (queue *Queue) Offer(value interface{}) error {
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

func (queue *Queue) DequeueTimer(timer *Timer, waitPool *ants.Pool) (value interface{}, hasValue bool) {
	waitPool.Submit(func() {
		timeFa := <-timer.C
		queue.lock.Lock()
		if timeFa {
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
	})
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
			fa := <-queue.flag
			if !fa {
				timer.Close()
				return nil, false
			}
		}
	}
}

func (queue *Queue) DequeueTimer2(timer *Timer, waitPool *ants.Pool) (value interface{}, hasValue bool) {
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

var poolQueue = &sync.Pool{
	New: func() interface{} {
		return &Queue{}
	},
}

func GetQueue() *Queue {
	queue := poolQueue.Get().(*Queue)
	queue.sliceQueue = GetSliceQueue()
	queue.lock = new(sync.RWMutex)
	queue.flag = make(chan bool)
	queue.waitNum = 0
	return queue
}
func FreeQueue(queue *Queue) {
	FreeSliceQueue(queue.sliceQueue)
	poolQueue.Put(queue)
}
