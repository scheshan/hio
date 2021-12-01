package hio

import "testing"

func TestMalloc(t *testing.T) {
	b := Malloc(24)

	if cap(b.buf) != 32 {
		t.Fail()
	}
}

func TestBuffer_Release(t *testing.T) {
	b := Malloc(24)
	b.Release()

	if b.ref != 0 {
		t.Fail()
	}
}
