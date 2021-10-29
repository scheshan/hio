package hio

import "sync"

var defaultMemoryPoolMaxSize = 32

type MemoryPool struct {
	maxSize int
	pools   []*sync.Pool
}

func (t *MemoryPool) Get(size int) []byte {
	if size <= 0 {
		panic("invalid size")
	}

	ind := t.log2(size)
	if ind >= t.maxSize {
		panic("request memory exceed maximum size")
	}

	data := t.pools[ind].Get().([]byte)
	return data[:cap(data)]
}

func (t *MemoryPool) Put(data []byte) {
	size := cap(data)
	if size&(size-1) != 0 {
		panic("invalid byte slice")
	}

	ind := t.log2(size)
	if ind >= t.maxSize {
		panic("request memory exceed maximum size")
	}

	t.pools[ind].Put(data)
}

func (t *MemoryPool) log2(size int) int {
	if size <= 0 {
		panic("invalid size")
	}

	s := size
	r := 0

	for s != 0 {
		s = s >> 1
		r++
	}

	if size&(size-1) != 0 {
		r++
	}

	return r - 1
}

func NewMemoryPool() *MemoryPool {
	return NewMemoryPoolSize(defaultMemoryPoolMaxSize)
}

func NewMemoryPoolSize(maxSize int) *MemoryPool {
	mp := &MemoryPool{}
	mp.maxSize = maxSize
	mp.pools = make([]*sync.Pool, maxSize, maxSize)

	for i := 0; i < maxSize; i++ {
		size := 1 << i
		mp.pools[i] = &sync.Pool{
			New: func() interface{} {
				return make([]byte, size, size)
			},
		}
	}

	return mp
}
