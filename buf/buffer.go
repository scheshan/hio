package buf

import (
	"errors"
	"sync/atomic"
)

var bufferMinNodeSize = 8192
var bufferMaxNodeSize = 2 << 20

var ErrBufferNoEnoughData = errors.New("buffer has no enough data to read")

type Buffer struct {
	head        *node
	tail        *node
	size        int
	minNodeSize int
	ref         int32
}

func (t *Buffer) addNewNode() {
	node := newNode()
	node.b = newBytes(t.minNodeSize)

	t.addTail(node)
}

func (t *Buffer) addNewNodeData(data []byte) {
	node := newNode()
	node.b = newBytes(len(data))
	copy(node.b, data)

	node.w = len(data)
	t.addTail(node)
}

func (t *Buffer) addTail(node *node) {
	if t.tail == nil {
		t.head = node
		t.tail = node
	} else {
		t.tail.next = node
		t.tail = node
	}
	t.size += node.w - node.r
}

func (t *Buffer) checkSize(size int) error {
	if t.size < size {
		return ErrBufferNoEnoughData
	}

	return nil
}

func (t *Buffer) checkWrite(size int) error {
	return nil
}

func (t *Buffer) readHeadBytes(n int) []byte {
	res := t.head.readBytes(n)

	t.skipHeadBytes(n)

	return res
}

func (t *Buffer) skipHeadBytes(n int) {
	t.head.r += n
	t.size -= n
	if t.head.readableBytes() == 0 {
		next := t.head.next
		t.head.release()
		t.head = next

		if t.head == nil {
			t.tail = nil
		}
	}
}

func (t *Buffer) ensureWritable() {
	if t.tail == nil || t.tail.writableBytes() == 0 {
		t.addNewNode()
	}
}

func (t *Buffer) writeByte(b byte) {
	t.ensureWritable()

	t.tail.writeByte(b)
	t.size++
}

func (t *Buffer) writeUInt16(n uint16) {
	t.ensureWritable()

	if t.tail != nil && t.tail.writableBytes() >= 2 {
		t.tail.writeByte(byte(n>>8), byte(n))
		t.size += 2
	} else {
		t.writeByte(byte(n >> 8))
		t.writeByte(byte(n))
	}
}

func (t *Buffer) writeUInt32(n uint32) {
	t.ensureWritable()

	if t.tail != nil && t.tail.writableBytes() >= 4 {
		t.tail.writeByte(byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
		t.size += 4
	} else {
		t.writeUInt16(uint16(n >> 16))
		t.writeUInt16(uint16(n))
	}
}

func (t *Buffer) writeUInt64(n uint64) {
	t.ensureWritable()

	if t.tail != nil && t.tail.writableBytes() >= 8 {
		t.tail.writeByte(byte(n>>56), byte(n>>48), byte(n>>40), byte(n>>32), byte(n>>24), byte(n>>16), byte(n>>8), byte(n))
		t.size += 8
	} else {
		t.writeUInt32(uint32(n >> 32))
		t.writeUInt32(uint32(n))
	}
}

func (t *Buffer) Release() {
	ref := atomic.AddInt32(&t.ref, -1)
	if ref != 0 {
		return
	}

	for t.head != nil {
		t.head.release()
		t.head = t.head.next
	}

	t.tail = nil
	t.size = 0
	t.minNodeSize = 0
	returnBuffer(t)
}

func (t *Buffer) IncrRef() {
	atomic.AddInt32(&t.ref, 1)
}
