package util

import (
	"context"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type element struct {
	next  *element
	value interface{}
}

func newElement(value interface{}) *element {
	return &element{value: value}
}

var poolElement = &sync.Pool{
	New: func() interface{} {
		return new(element)
	},
}

func getElement(value interface{}) *element {
	ele := poolElement.Get().(*element)
	ele.value = value
	return ele
}
func freeElement(ele *element) {
	ele.next = nil
	poolElement.Put(ele)
}

type ele struct {
	next   *ele
	fq     []interface{}
	wIndex int32
	rIndex int32
	cap    int32
	lCap   int32
}

func (q *ele) insert(value interface{}) int32 {
	num := atomic.AddInt32(&q.wIndex, 1)
	if num > q.lCap {
		return num
	}
	q.fq[num] = value
	return num
}
func (q *ele) get() (interface{}, int32) {
	num := atomic.AddInt32(&q.rIndex, 1)
	if num > q.lCap {
		return nil, num
	}
	for {
		n := q.fq[num]
		if n != nil {
			q.fq[num] = nil
			return n, num
		} else {
			runtime.Gosched()
		}
	}
}

func (q *ele) read(num int32) interface{} {
	return q.fq[num]
}

var poolEle = &sync.Pool{
	New: func() interface{} {
		return new(ele)
	},
}

func newEle(ca int32) *ele {
	ele := poolEle.Get().(*ele)
	ele.fq = make([]interface{}, ca)
	ele.cap = ca
	ele.wIndex = -1
	ele.rIndex = -1
	ele.lCap = ca - 1
	return ele
}
func freeEle(ele *ele) {
	ele.next = nil
	poolEle.Put(ele)
}

type timer struct {
	t     *time.Timer
	isEnd chan bool
}

func newTimer() *timer {
	return &timer{t: time.NewTimer(time.Second * 10), isEnd: make(chan bool)}
}
func (timer *timer) wait() bool {
	select {
	case <-timer.t.C:
		{
			return false
		}
	case fa := <-timer.isEnd:
		{
			return fa
		}
	}
	return false
}
func (timer *timer) end() {
	fa := timer.t.Stop()
	if fa {
		timer.isEnd <- true
	}
}
func (timer *timer) reset(duration time.Duration) {
	timer.t.Reset(duration)
}

var poolTimer = &sync.Pool{
	New: func() interface{} {
		return newTimer()
	},
}

func getTimer(duration time.Duration) *timer {
	ti := poolTimer.Get().(*timer)
	ti.reset(duration)
	return ti
}
func freeTimer(timer *timer) {
	poolTimer.Put(timer)
}

type operate struct {
	ctx context.Context
}

func newOperate(ctx context.Context) *operate {
	return &operate{ctx: ctx}
}
func (op *operate) wait() bool {
	select {
	case <-op.ctx.Done():
		return true
	}
	return false
}
func (op *operate) isClose() bool {
	return op.ctx.Err() != nil
}

var poolOperate = &sync.Pool{
	New: func() interface{} {
		return &operate{}
	},
}

func getOperate(ctx context.Context) *operate {
	v, _ := poolOperate.Get().(*operate)
	v.ctx = ctx
	return v
}
func freeOperate(op *operate) {
	poolOperate.Put(op)
}

type VQueue struct {
	write   *ele
	read    *ele
	num     int32
	waitNum int32
	flag    chan bool
	cap     int32
	lCap    int32
	lock    *sync.Mutex
}

func (queue *VQueue) Offer(value interface{}) (nu int32) {

	for {
		num := queue.write.insert(value)
		if num < queue.cap {
			nu = atomic.AddInt32(&queue.num, 1)
			queue.lock.Lock()
			if atomic.LoadInt32(&queue.waitNum) > 0 {
				atomic.AddInt32(&queue.waitNum, -1)
				queue.lock.Unlock()
				queue.flag <- true
			} else {
				queue.lock.Unlock()
			}
			return
		} else {
			if num == queue.cap {
				queue.write.next = newEle(queue.cap)
				queue.write = queue.write.next
			} else {
				runtime.Gosched()
			}
		}
	}
}
func (queue *VQueue) poll() (value interface{}, nu int32, hasValue bool) {
	v, num := queue.read.get()
	if num < queue.cap {
		nu = atomic.AddInt32(&queue.num, -1)
		return v, nu, true
	} else {
		if num == queue.cap {
			r := queue.read
			for {
				if queue.read.next != nil {
					queue.read = queue.read.next
					break
				} else {
					runtime.Gosched()
				}
			}
			freeEle(r)
		} else {
			runtime.Gosched()
		}
	}
	return
}

func (queue *VQueue) Poll() (value interface{}, nu int32) {
	for {
		queue.lock.Lock()
		if atomic.LoadInt32(&queue.num) == 0 {
			atomic.AddInt32(&queue.waitNum, 1)
			queue.lock.Unlock()
			<-queue.flag
		} else {
			queue.lock.Unlock()
			v, num, has := queue.poll()
			if has {
				return v, num
			}
		}
	}
}

func (queue *VQueue) Dequeue(ctx context.Context) (value interface{}, num int32, hasClose bool) {
	for {
		queue.lock.Lock()
		if atomic.LoadInt32(&queue.num) == 0 {
			atomic.AddInt32(&queue.waitNum, 1)
			queue.lock.Unlock()
			var op = getOperate(ctx)
			go func() {
				fa := op.wait()
				queue.lock.Lock()
				if atomic.LoadInt32(&queue.waitNum) > 0 {
					atomic.AddInt32(&queue.waitNum, -1)
					queue.lock.Unlock()
					queue.flag <- !fa
				} else {
					queue.lock.Unlock()
				}
			}()
			flag := <-queue.flag
			freeOperate(op)
			if !flag {
				return nil, 0, true
			}
		} else {
			queue.lock.Unlock()
			v, num, has := queue.poll()
			if has {
				return v, num, false
			}
		}
	}
}
func (queue *VQueue) Num() int32 {
	return atomic.LoadInt32(&queue.num)
}
func (queue *VQueue) Take(duration time.Duration) (value interface{}, num int32) {
	for {
		queue.lock.Lock()
		if atomic.LoadInt32(&queue.num) == 0 {
			atomic.AddInt32(&queue.waitNum, 1)
			queue.lock.Unlock()
			tm := getTimer(duration)
			go func() {
				fa := tm.wait()
				if !fa {
					queue.lock.Lock()
					if atomic.LoadInt32(&queue.waitNum) > 0 {
						atomic.AddInt32(&queue.waitNum, -1)
						queue.lock.Unlock()
						queue.flag <- false
					} else {
						queue.lock.Unlock()
					}
				}
			}()
			flag := <-queue.flag
			tm.end()
			freeTimer(tm)
			if !flag {
				return nil, 0
			}
		} else {
			queue.lock.Unlock()
			v, num, has := queue.poll()
			if has {
				return v, num
			}
		}
	}
}

func NewVQueue() *VQueue {
	var ca int32 = 128
	el := newEle(ca)
	return &VQueue{write: el, read: el, flag: make(chan bool), num: 0, cap: ca, lock: new(sync.Mutex)}
}

type Queue struct {
	input   *element
	output  *element
	ch      chan bool
	waitNum int32
	num     int32
	lock    *sync.RWMutex
	rLock   *sync.Mutex
	timer   *timer
}

func NewQueue() *Queue {
	return &Queue{ch: make(chan bool), waitNum: 0, num: 0, lock: new(sync.RWMutex), rLock: new(sync.Mutex)}
}
func (queue *Queue) Offer(value interface{}) (num int32) {
	ele := getElement(value)
	queue.lock.Lock()
	if atomic.CompareAndSwapInt32(&queue.num, 0, 1) {
		queue.input = ele
		queue.output = ele
		num = 1
	} else {
		queue.input.next = ele
		queue.input = ele
		num = atomic.AddInt32(&queue.num, 1)
	}
	if queue.waitNum > 0 {
		atomic.AddInt32(&queue.waitNum, -1)
		queue.lock.Unlock()
		queue.ch <- true
	} else {
		queue.lock.Unlock()
	}
	return
}
func (queue *Queue) Num() int32 {
	return queue.num
}
func (queue *Queue) Poll() (value interface{}, num int32) {
	for {
		queue.lock.Lock()
		if queue.num > 0 {
			if queue.num == 1 {
				value, num = queue.readOne()
				queue.lock.Unlock()
				return
			} else {
				queue.lock.Unlock()
				queue.rLock.Lock()
				val, n, last := queue.readGtOne()
				if last {
					queue.rLock.Unlock()
				} else {
					queue.rLock.Unlock()
					return val, n
				}
			}
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			<-queue.ch
		}
	}
}
func (queue *Queue) Peek() (value interface{}, num int32) {
	queue.lock.RLock()
	num = queue.num
	if queue.num > 0 {
		value = queue.output.value
		queue.lock.RUnlock()
		return
	} else {
		queue.lock.RUnlock()
		return nil, 0
	}
}
func (queue *Queue) readOne() (value interface{}, num int32) {
	var ele = queue.output
	value = ele.value
	num = atomic.AddInt32(&queue.num, -1)
	freeElement(ele)
	return value, num
}
func (queue *Queue) readGtOne() (value interface{}, num int32, isLast bool) {
	var ele = queue.output
	if ele.next == nil {
		return nil, 0, true
	}
	value = ele.value
	queue.output = ele.next
	num = atomic.AddInt32(&queue.num, -1)
	freeElement(ele)
	return value, num, false
}

func (queue *Queue) Dequeue(ctx context.Context) (value interface{}, num int32, cols bool) {
	var hasReturn = false
	for {
		queue.lock.Lock()
		if queue.num > 0 {
			if queue.num == 1 {
				value, num = queue.readOne()
				queue.lock.Unlock()
				hasReturn = true
				return value, num, false
			} else {
				queue.lock.Unlock()
				queue.rLock.Lock()
				val, n, last := queue.readGtOne()
				if last {
					queue.rLock.Unlock()
				} else {
					queue.rLock.Unlock()
					hasReturn = true
					return val, n, false
				}
			}
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			var op = getOperate(ctx)
			go func() {
				fa := op.wait()
				if hasReturn {
					return
				}
				queue.lock.Lock()
				if queue.waitNum > 0 {
					queue.waitNum--
					queue.lock.Unlock()
					queue.ch <- !fa
				} else {
					queue.lock.Unlock()
				}
			}()
			flag := <-queue.ch
			freeOperate(op)
			if !flag {
				hasReturn = true
				return nil, 0, true
			}
		}
	}
}

func (queue *Queue) Take(duration time.Duration) (value interface{}, num int32) {
	for {
		queue.lock.Lock()
		if queue.num > 0 {
			if queue.num == 1 {
				value, num = queue.readOne()
				queue.lock.Unlock()
				return
			} else {
				queue.lock.Unlock()
				queue.rLock.Lock()
				val, n, last := queue.readGtOne()
				if last {
					queue.rLock.Unlock()
				} else {
					queue.rLock.Unlock()
					return val, n
				}
			}
		} else {
			queue.waitNum++
			queue.lock.Unlock()
			tm := getTimer(duration)
			go func() {
				fa := tm.wait()
				if !fa {
					queue.lock.Lock()
					if queue.waitNum > 0 {
						queue.waitNum--
						queue.lock.Unlock()
						queue.ch <- false
					} else {
						queue.lock.Unlock()
					}
				}
			}()
			flag := <-queue.ch
			tm.end()
			freeTimer(tm)
			if !flag {
				return nil, 0
			}
		}
	}
}
