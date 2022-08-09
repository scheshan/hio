package hio

import "math/rand"

type LoadBalancerType int

const (
	RoundRobin LoadBalancerType = 0
	Random     LoadBalancerType = 1
	//LeastConnection LoadBalancerType = 2
)

type loadBalancer interface {
	Choose(loops []*eventLoop) *eventLoop
}

type lbRoundRobin struct {
	idx int
}

func (t *lbRoundRobin) Choose(loops []*eventLoop) *eventLoop {
	l := loops[t.idx]

	t.idx++
	if t.idx >= len(loops) {
		t.idx = 0
	}

	return l
}

type lbRandom struct {
	rnd rand.Rand
}

func (t *lbRandom) Choose(loops []*eventLoop) *eventLoop {
	idx := t.rnd.Intn(len(loops))
	return loops[idx]
}
