package util

import (
	"sync"
	"sync/atomic"
)

type waitNumGroup struct {
	waitGroup *sync.WaitGroup
	num       uint32
}

func (g *waitNumGroup) AddOne() {
	atomic.AddUint32(&g.num, 1)
	g.waitGroup.Add(1)
}

func (g *waitNumGroup) Done() {
	if atomic.AddUint32(&g.num, -1) >= 0 {
		g.waitGroup.Done()
	}
}

func (g *waitNumGroup) Wait() {
	g.waitGroup.Wait()
}

func NewWaitNumGroup() *waitNumGroup {
	return &waitNumGroup{waitGroup: new(sync.WaitGroup)}
}
