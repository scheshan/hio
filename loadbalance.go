package hio

import (
	"math/rand"
	"sync/atomic"
)

type LoadBalancer interface {
	Choose(loops []*EventLoop) *EventLoop
}

type lbRoundRobin struct {
	idx int32
}

func (t *lbRoundRobin) Choose(loops []*EventLoop) *EventLoop {
	var loop *EventLoop

	for loop == nil {
		idx := t.idx + 1
		if idx >= int32(len(loops)) {
			idx = 0
		}
		if atomic.CompareAndSwapInt32(&t.idx, t.idx, idx) {
			loop = loops[idx]
		}
	}

	return loop
}

type lbRandom struct {
}

func (t *lbRandom) Choose(loops []*EventLoop) *EventLoop {
	idx := rand.Intn(len(loops))
	return loops[idx]
}
