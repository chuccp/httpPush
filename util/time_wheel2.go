package util

import (
	"context"
	"sync"
	"time"
)

type Handle func(value ...any)

type TimeWheel2 struct {
	tick            int32
	bucketsNum      int32
	bucketsMaxIndex int32
	readerIndex     int32
	buckets         []*bucket2
	data            map[string]int32
	lock            *sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
}

type bucket2 struct {
	data *sync.Map
	lock *sync.Mutex
}
type handle struct {
	handle Handle
	value  []any
}

func NewTimeWheel2(tickSeconds int32, bucketsNum int32) *TimeWheel2 {
	timeWheel := &TimeWheel2{tick: tickSeconds, bucketsNum: bucketsNum, bucketsMaxIndex: bucketsNum - 1, data: make(map[string]int32), lock: new(sync.RWMutex)}
	timeWheel.ctx, timeWheel.cancel = context.WithCancel(context.Background())
	timeWheel.buckets = make([]*bucket2, bucketsNum)
	for i := 0; i < int(bucketsNum); i++ {
		timeWheel.buckets[i] = &bucket2{data: new(sync.Map), lock: new(sync.Mutex)}
	}
	return timeWheel
}
func (tw *TimeWheel2) addHandle(index int32, id string, f Handle, value ...any) {
	tw.lock.Lock()
	defer tw.lock.Unlock()
	v, ok := tw.data[id]
	if ok {
		tw.buckets[v].data.Delete(id)
	}
	tw.buckets[index].data.Store(id, &handle{handle: f, value: value})
	tw.data[id] = index
}
func (tw *TimeWheel2) AfterFunc(tickSeconds int32, id string, f Handle, value ...any) {
	index := tickSeconds / tw.tick
	y := tickSeconds % tw.tick
	if y > 0 {
		index = index + 1
	}
	readerIndex := tw.readerIndex
	vIndex := index + readerIndex
	if vIndex >= tw.bucketsNum {
		vIndex = vIndex - tw.bucketsNum
	}
	tw.addHandle(vIndex, id, f, value...)
}
func (tw *TimeWheel2) DeleteFunc(id string) {
	tw.lock.Lock()
	defer tw.lock.Unlock()
	v, ok := tw.data[id]
	if ok {
		tw.buckets[v].data.Delete(id)
	}
	delete(tw.data, id)
}

func (tw *TimeWheel2) getBucketsByIndex(index int32) *bucket2 {
	return tw.buckets[index]
}

func (tw *TimeWheel2) scheduler() {
	index := tw.readerIndex
	db := tw.getBucketsByIndex(index)
	db.data.Range(func(key, value any) bool {
		tw.lock.Lock()
		db.data.Delete(key)
		k := key.(string)
		i, ok := tw.data[k]
		if ok && i == index {
			delete(tw.data, k)
		}
		tw.lock.Unlock()
		handel, ok := value.(*handle)
		if ok {
			handel.handle(handel.value...)
		}
		return true
	})
	if tw.readerIndex >= tw.bucketsMaxIndex {
		tw.readerIndex = 0
	} else {
		tw.readerIndex++
	}
}
func (tw *TimeWheel2) Stop() {
	tw.cancel()
}

func (tw *TimeWheel2) Start() {
	ticker := time.NewTicker(time.Duration(tw.tick) * time.Second)
	for {
		select {
		case <-ticker.C:
			tw.scheduler()
		case <-tw.ctx.Done():
			ticker.Stop()
			return
		}
	}
}
