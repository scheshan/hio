package hio

import "testing"

func Test_ind(t *testing.T) {
	p := NewBytesPool()

	assertInd(t, 0, p.ind(1))
	assertInd(t, 1, p.ind(2))
	assertInd(t, 2, p.ind(3))
	assertInd(t, 2, p.ind(4))
	assertInd(t, 3, p.ind(5))
	assertInd(t, 3, p.ind(7))
	assertInd(t, 3, p.ind(8))
}

func assertInd(t *testing.T, except int, fact int) {
	if except != fact {
		t.Errorf("except: %v, fact: %v", except, fact)
	}
}

func Test_Get(t *testing.T) {
	p := defaultBytesPool

	b := p.get(20)

	if cap(b.buf) != 32 {
		t.Fail()
	}
}
