package util

import (
	"context"
	"github.com/rfyiamcool/go-timewheel"
	"sync"
	"sync/atomic"
)

type CancelContext struct {
	ctx           context.Context
	ctxCancelFunc context.CancelFunc
	close         *atomic.Bool
	once          *sync.Once
}

func NewCancelContext() *CancelContext {
	cc := &CancelContext{once: new(sync.Once)}
	cc.ctx, cc.ctxCancelFunc = context.WithCancel(context.Background())
	cc.close = new(atomic.Bool)
	cc.close.Store(false)
	return cc
}

var CloseExceeded error = closeExceededError{}

type closeExceededError struct{}

func (closeExceededError) Error() string { return "context close exceeded" }
func (c *CancelContext) Wait() {
	<-c.ctx.Done()
}
func (c *CancelContext) Cancel() {
	if !c.close.Load() {
		c.cancel()
	}
}

func (c *CancelContext) cancel() {
	c.once.Do(c.ctxCancelFunc)
}

func (c *CancelContext) Close() {
	if c.close.CompareAndSwap(false, true) {
		c.cancel()
	}
}
func (c *CancelContext) Err() error {
	if c.close.Load() {
		return CloseExceeded
	}
	return c.ctx.Err()
}

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
