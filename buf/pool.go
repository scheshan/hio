package buf

import "sync"

var nodePool = &sync.Pool{
	New: func() interface{} {
		return &node{}
	},
}

var bufferPool = &sync.Pool{
	New: func() interface{} {
		return &Buffer{}
	},
}

var bPool = newBytesPoolSize(64)

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

func (t *bytesPool) get(size int) []byte {
	if size <= 0 || size > t.maxSize {
		return nil
	}

	ind := t.ind(size)
	b := t.pools[ind].Get().([]byte)

	return b
}

func (t *bytesPool) put(data []byte) {
	if t.isPowOf2(len(data)) {
		return
	}

	ind := t.ind(len(data))
	if ind >= len(t.pools) {
		return
	}

	t.pools[ind].Put(data)
}

func newBytesPoolSize(size int) *bytesPool {
	if size <= 0 || size > 64 {
		panic("invalid pool size")
	}

	p := new(bytesPool)
	p.maxSize = 1 << (size - 1)
	p.pools = make([]*sync.Pool, size, size)
	for i := 0; i < size; i++ {
		bytes := 1 << i
		p.pools[i] = &sync.Pool{
			New: func() interface{} {
				buf := make([]byte, bytes)

				return buf
			},
		}
	}

	return p
}

func newBytes(size int) []byte {
	if size > bPool.maxSize {
		return make([]byte, size)
	}

	return bPool.get(size)
}

func returnBytes(data []byte) {
	bPool.put(data)
}

func newNode() *node {
	return nodePool.Get().(*node)
}

func returnNode(n *node) {
	nodePool.Put(n)
}

func NewBuffer() *Buffer {
	return bufferPool.Get().(*Buffer)
}

func NewBufferSize(size int) *Buffer {
	if size <= 0 {
		size = bufferMinNodeSize
	}

	if size > bufferMaxNodeSize {
		size = bufferMaxNodeSize
	}

	buf := bufferPool.Get().(*Buffer)
	buf.minNodeSize = size

	return buf
}

func returnBuffer(b *Buffer) {
	bufferPool.Put(b)
}
