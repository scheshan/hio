package hio

import (
	"sync"
)

var defaultBytesPoolSize = 46
var defaultBytesPool = NewBytesPool()

type bytesPool struct {
	pools   []*sync.Pool
	maxSize int
}

func (t *bytesPool) ind(size int) int {
	ind := 0
	if !t.isPowOf2(size) {
		ind = 1
	}

	for size > 0 {
		size = size >> 1
		ind++
	}

	return ind - 1
}

func (t *bytesPool) isPowOf2(size int) bool {
	return size&(size-1) == 0
}

func (t *bytesPool) get(size int) *Bytes {
	if size <= 0 || size > t.maxSize {
		return nil
	}

	ind := t.ind(size)
	data := t.pools[ind].Get().(*Bytes)

	return data
}

func (t *bytesPool) put(b *Bytes) {
	t.pools[b.ind].Put(b)
}

func NewBytesPool() *bytesPool {
	return NewBytesPoolSize(defaultBytesPoolSize)
}

func NewBytesPoolSize(size int) *bytesPool {
	if size <= 0 || size > 64 {
		panic("invalid pool size")
	}

	p := new(bytesPool)
	p.maxSize = 1 << (size - 1)
	p.pools = make([]*sync.Pool, size, size)
	for i := 0; i < size; i++ {
		bytes := 1 << i
		ind := i
		p.pools[i] = &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, bytes)
				b := &Bytes{
					buf: buf,
					ind: ind,
				}

				return b
			},
		}
	}

	return p
}
