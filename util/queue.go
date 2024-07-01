package util

import (
	"github.com/rfyiamcool/go-timewheel"
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

func (queue *Queue) DequeueTimer(timer *timewheel.Timer) (value interface{}, hasValue bool) {
	go func() {
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
	}()
	for {
		queue.lock.Lock()
		v, err := queue.sliceQueue.Read()
		if err == nil {
			queue.lock.Unlock()
			timer.Stop()
			close(timer.C)
			return v, true
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			fa := <-queue.flag
			if !fa {
				timer.Stop()
				close(timer.C)
				return nil, false
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
