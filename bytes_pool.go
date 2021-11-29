package hio

import (
	"errors"
	"sync"
)

var defaultBytesPoolSize = 46
var defaultBytesPool = newBytesPool(defaultBytesPoolSize)

type BytesPool struct {
	pools   []*sync.Pool
	maxSize int
}

func (t *BytesPool) ind(size int) int {
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

func (t *BytesPool) Get(size int) (*Bytes, error) {
	if size <= 0 {
		return nil, errors.New("invalid size")
	}
	if size > t.maxSize {
		return nil, errors.New("request size too large")
	}

	ind := t.ind(size)
	data := t.pools[ind].Get().(*Bytes)

	return data, nil
}

func (t *BytesPool) isPowOf2(size int) bool {
	return size&(size-1) == 0
}

func (t *BytesPool) put(b *Bytes) {
	if b.p != t {
		return
	}

	t.pools[b.ind].Put(b)
}

func newBytesPool(size int) *BytesPool {
	if size <= 0 || size > 64 {
		panic("invalid pool size")
	}

	p := new(BytesPool)
	p.maxSize = 1 << (size - 1)
	p.pools = make([]*sync.Pool, size, size)
	for i := 0; i < size; i++ {
		bytes := 1 << i
		ind := i
		p.pools[i] = &sync.Pool{
			New: func() interface{} {
				d := make([]byte, bytes, bytes)
				b := &Bytes{
					b:   d,
					ind: ind,
					p:   p,
				}
				return b
			},
		}
	}

	return p
}
