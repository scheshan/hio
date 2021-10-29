package hio

import (
	"testing"
)

func TestMemoryPool_log2(t *testing.T) {
	mp := NewMemoryPool()

	for i := 0; i < defaultMemoryPoolMaxSize; i++ {
		if mp.log2(1<<i) != i {
			t.Fatalf("log2 %v failed", 1<<i)
		}
	}

	for i := 0; i < defaultMemoryPoolMaxSize-1; i++ {
		if mp.log2((1<<i)+1) != i+1 {
			t.Fatalf("log2 %v failed", (1<<i)+1)
		}
	}

	for i := 2; i < defaultMemoryPoolMaxSize; i++ {
		if mp.log2((1<<i)-1) != i {
			t.Fatalf("log2 %v failed", (1<<i)-1)
		}
	}
}

func TestMemoryPool_Get(t *testing.T) {
	mp := NewMemoryPool()

	for i := 0; i < defaultMemoryPoolMaxSize>>1; i++ {
		size := 1 << i
		data := mp.Get(size)
		if cap(data) != size {
			t.Fatalf("Get return wrong size. request: %v, return: %v", size, cap(data))
		}
	}
}
