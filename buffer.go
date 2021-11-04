package hio

import (
	"errors"
)

var ErrBufferEmpty = errors.New("buffer is empty")
var ErrBufferOutOfRange = errors.New("buffer out of range")

type Buffer struct {
	head *bufferNode
	wi   *bufferNode
	ri   *bufferNode
	mp   *MemoryPool
}

func (t *Buffer) Read(data []byte) (n int, err error) {
	for t.CanRead() {
		cn := copy(data[n:], t.ri.buf[t.ri.ri:t.ri.wi])

		n += cn
		t.Skip(cn)

		if n >= len(data) {
			return
		}
	}

	return
}

func (t *Buffer) ReadByte() (b byte, err error) {
	if !t.CanRead() {
		return 0, ErrBufferEmpty
	}

	b = t.ri.buf[t.ri.ri]
	t.Skip(1)

	return b, nil
}

func (t *Buffer) Write(data []byte) (n int, err error) {
	i := 0

	if t.wi == nil {
		t.alloc(len(data))
	}

	for i < len(data) {
		cn := copy(t.wi.buf[t.wi.wi:], data[i:])
		i += cn
		t.wi.wi += cn

		if len(t.wi.buf) == cap(t.wi.buf) && i < len(data) {
			t.alloc(len(data) - i)
		}
	}

	return 0, nil
}

func (t *Buffer) CanRead() bool {
	if t.ri == nil {
		return false
	}

	if t.ri == t.wi {
		if t.ri.ri == t.ri.wi {
			return false
		}
	}

	return true
}

func (t *Buffer) Skip(n int) error {
	for n > 0 {
		if t.ri == nil {
			return ErrBufferOutOfRange
		}

		available := t.ri.wi - t.ri.ri
		skipNum := n
		if skipNum > available {
			skipNum = available
		}

		t.ri.ri += skipNum
		n -= skipNum

		if t.ri.ri == t.ri.wi && t.ri != t.wi {
			t.ri = t.ri.next
		}

		if n > 0 && t.ri.next == nil {
			return ErrBufferOutOfRange
		}
	}

	return nil
}

func (t *Buffer) alloc(size int) {
	if size < 4096 {
		size = 4096
	}

	buf := t.mp.Get(size)
	node := new(bufferNode)
	node.buf = buf

	if t.wi == nil {
		t.head = node
		t.ri = node
	} else {
		t.wi.next = node
		node.prev = t.wi
	}

	if t.ri == t.wi && t.ri.ri == t.ri.wi {
		t.ri = node
	}
	t.wi = node
}

type bufferNode struct {
	buf  []byte
	prev *bufferNode
	next *bufferNode
	ri   int
	wi   int
}
