package hio

import (
	"math/rand"
	"sync/atomic"
)

type LoadBalancer interface {
	Choose(loops []*EventLoop) *EventLoop
}

type LoadBalancerRoundRobin struct {
	idx int32
}

func (t *LoadBalancerRoundRobin) Choose(loops []*EventLoop) *EventLoop {
	for {
		o := atomic.LoadInt32(&t.idx)
		var n = int(o) + 1

		if n >= len(loops) {
			n = 0
		}

		if atomic.CompareAndSwapInt32(&t.idx, o, int32(n)) {
			return loops[n]
		}
	}
}

type LoadBalancerRandom struct {
}

func (t *LoadBalancerRandom) Choose(loops []*EventLoop) *EventLoop {
	idx := rand.Intn(len(loops))

	return loops[idx]
}
