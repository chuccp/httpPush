package util

import (
	"context"
	"errors"
	"sync"
)

type Queue struct {
	sliceQueue *SliceQueue
	lock       *sync.RWMutex
	waitNum    int32
	flag       chan bool
}

func NewQueue() *Queue {
	return &Queue{sliceQueue: new(SliceQueue), lock: new(sync.RWMutex), flag: make(chan bool)}
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
func (queue *Queue) Dequeue(ctx context.Context) (value interface{}, err error, hasClose bool) {

	for {
		queue.lock.Lock()
		num := queue.sliceQueue.Len()
		if num > 0 {
			v, err := queue.sliceQueue.Read()
			queue.lock.Unlock()
			return v, err, false
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
				return nil, nil, true
			}
		}
	}

}
func (queue *Queue) Poll() (value interface{}, err error) {
	for {
		queue.lock.Lock()
		num := queue.sliceQueue.Len()
		if num > 0 {
			ele, err := queue.sliceQueue.Read()
			queue.lock.Unlock()
			return ele, err
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			<-queue.flag
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
