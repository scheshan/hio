package hio

import "testing"

func Test_ind(t *testing.T) {
	p := defaultBytesPool

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

	b, err := p.Get(20)
	if err != nil {
		t.Fatal(err)
	}

	if cap(b.b) != 32 {
		t.Fail()
	}
}
