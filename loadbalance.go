package hio

import (
	"math/rand"
	"sync/atomic"
)

type LoadBalancer interface {
	Initialize(loops []*EventLoop)
	Choose() *EventLoop
}

type lbRoundRobin struct {
	idx   int32
	loops []*EventLoop
}

func (t *lbRoundRobin) Initialize(loops []*EventLoop) {
	t.loops = loops
}

func (t *lbRoundRobin) Choose() *EventLoop {
	var loop *EventLoop
	loops := t.loops

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
	loops []*EventLoop
}

func (t *lbRandom) Initialize(loops []*EventLoop) {
	t.loops = loops
}

func (t *lbRandom) Choose() *EventLoop {
	loops := t.loops

	idx := rand.Intn(len(loops))
	return loops[idx]
}
