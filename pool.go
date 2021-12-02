package hio

import "sync"

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

type defaultPool struct {
	bytes       *bytesPool
	buffers     *sync.Pool
	bufferNodes *sync.Pool
}

func (t *defaultPool) getBuffer() *Buffer {
	return t.buffers.Get().(*Buffer)
}

func (t *defaultPool) putBuffer(buf *Buffer) {
	t.buffers.Put(buf)
}

func (t *defaultPool) getNode(size int) *bufferNode {
	node := t.bufferNodes.Get().(*bufferNode)
	node.ref = 1
	node.b = t.getBytes(size)

	return node
}

func (t *defaultPool) putNode(node *bufferNode) {
	t.bufferNodes.Put(node)
}

func (t *defaultPool) getBytes(size int) []byte {
	return t.bytes.get(size)
}

func (t *defaultPool) putBytes(data []byte) {
	t.bytes.put(data)
}

func newPool() *defaultPool {
	p := &defaultPool{}
	p.buffers = &sync.Pool{
		New: func() interface{} {
			buf := &Buffer{}
			return buf
		},
	}
	p.bufferNodes = &sync.Pool{
		New: func() interface{} {
			node := &bufferNode{}
			return node
		},
	}
	p.bytes = newBytesPoolSize(46)

	return p
}

var pool *defaultPool = newPool()
