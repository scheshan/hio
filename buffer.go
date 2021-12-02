package hio

import "errors"

var ErrBufferNoEnoughData = errors.New("buffer has no enough data to read")

type Buffer struct {
	head *bufferNode
	tail *bufferNode
	r    *bufferNode
	size int
}

func (t *Buffer) Write(data []byte) {
	if t.tail == nil {
		t.addNewNode(data)
		return
	}

	n := t.tail.copy(data)

	t.size += n
	if n == len(data) {
		return
	}

	t.addNewNode(data[n:])
}

func (t *Buffer) CanRead() bool {
	return t.ReadableBytes() > 0
}

func (t *Buffer) ReadableBytes() int {
	return t.size
}

func (t *Buffer) ReadByte() (byte, error) {
	if t.ReadableBytes() < 1 {
		return 0, ErrBufferNoEnoughData
	}

	for {
		if t.r.readableBytes() > 0 {
			return t.r.nextByte(), nil
		}

		t.r = t.r.next
	}
}

func (t *Buffer) ReadBytes(n int) ([]byte, error) {
	if t.ReadableBytes() < n {
		return nil, ErrBufferNoEnoughData
	}

	res := make([]byte, n)

	cnt := 0
	for cnt < n {
		if t.r.readableBytes() > 0 {
			rn := n - cnt
			if rn > t.r.readableBytes() {
				rn = t.r.readableBytes()
			}

			b := t.r.nextBytes(rn)
			copy(res[cnt:], b)

			cnt += rn
		}
	}

	return res, nil
}

func (t *Buffer) Release() {
	for t.head != nil {
		n := t.head.next
		t.head.next = nil
		t.head.release()

		t.head = n
	}
}

func (t *Buffer) addNewNode(data []byte) {
	size := len(data)
	node := pool.getNode(size)
	node.copy(data)

	if t.tail == nil {
		t.head = node
		t.tail = node
		t.r = node
	} else {
		t.tail.next = node
		t.tail = node
	}

	t.tail.w = size
	t.size += size
}
