package util

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

const logNum = 10

// TimeWheel 单圈时间轮，只用于特定场景  ，timer 最大时间不能超过 tick*bucketsNum
type TimeWheel struct {
	tick            int32
	bucketsNum      int32
	bucketsMaxIndex int32
	readerIndex     int32
	buckets         []*bucket
	// 上下文
	ctx context.Context

	// 取消函数
	cancel context.CancelFunc

	timeWheelLog []*TimeWheelLog
	logIndex     int
}

type TimeWheelLog struct {
	Num       int
	StartTime *time.Time
	EndTime   *time.Time
}

type bucket struct {
	queue *SliceQueue
	lock  *sync.Mutex
}

func (b *bucket) add(timer *Timer) {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.queue.Write(timer)
}
func (tw *TimeWheel) addTimer(index int32, timer *Timer) {
	tw.buckets[index].add(timer)
}
func (b *bucket) read() (*Timer, error) {
	b.lock.Lock()
	defer b.lock.Unlock()
	read, err := b.queue.Read()
	if err != nil {
		return nil, err
	}
	return read.(*Timer), nil
}
func (b *bucket) len() int {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.queue.Len()
}

type Timer struct {
	C       <-chan bool
	c       chan<- bool
	isClose int32
}

func (t *Timer) run() {
	if atomic.CompareAndSwapInt32(&t.isClose, 0, 1) {
		t.c <- true
	}
}

func (t *Timer) Close() {
	atomic.StoreInt32(&t.isClose, 1)
	close(t.c)
}
func (tw *TimeWheel) GetLog() []*TimeWheelLog {
	return tw.timeWheelLog
}
func (tw *TimeWheel) NewTimer(tickSeconds int32) *Timer {
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
	c := make(chan bool, 1)
	timer := &Timer{C: c, c: c, isClose: 0}
	tw.addTimer(vIndex, timer)
	return timer
}
func (tw *TimeWheel) getBucketsByIndex(index int32) *bucket {
	return tw.buckets[index]
}

func (tw *TimeWheel) addLog(num int, startTime *time.Time, endTime *time.Time) {
	if tw.logIndex >= logNum {
		tw.logIndex = 0
	}
	tw.timeWheelLog[tw.logIndex] = &TimeWheelLog{Num: num, StartTime: startTime, EndTime: endTime}
	tw.logIndex++
}

func (tw *TimeWheel) scheduler() {
	index := tw.readerIndex
	sq := tw.getBucketsByIndex(index)
	startTime := time.Now()
	num := sq.len()
	for {
		tm, err := sq.read()
		if err != nil {
			if tw.readerIndex >= tw.bucketsMaxIndex {
				tw.readerIndex = 0
			} else {
				tw.readerIndex++
			}
			break
		} else {
			tm.run()
		}
	}
	if num > 0 {
		endTime := time.Now()
		tw.addLog(num, &startTime, &endTime)
	}
}
func (tw *TimeWheel) run() {
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
func (tw *TimeWheel) Stop() {
	tw.cancel()
}

// NewTimeWheel TimeWheel 单圈时间轮，只用于特定场景  ，timer 最大时间不能超过 tick*bucketsNum
func NewTimeWheel(tickSeconds int32, bucketsNum int32) *TimeWheel {
	timeWheel := &TimeWheel{tick: tickSeconds, bucketsNum: bucketsNum, bucketsMaxIndex: bucketsNum - 1}
	timeWheel.ctx, timeWheel.cancel = context.WithCancel(context.Background())
	timeWheel.buckets = make([]*bucket, bucketsNum)
	for i := 0; i < int(bucketsNum); i++ {
		timeWheel.buckets[i] = &bucket{queue: new(SliceQueue), lock: new(sync.Mutex)}
	}
	timeWheel.timeWheelLog = make([]*TimeWheelLog, logNum)
	timeWheel.logIndex = 0
	go timeWheel.run()
	return timeWheel
}
