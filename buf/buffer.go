package buf

import "errors"

var bufferMinNodeSize = 8192
var bufferMaxNodeSize = 2 << 20

var ErrBufferNoEnoughData = errors.New("buffer has no enough data to read")

type node struct {
	b    []byte
	next *node
	r    int
	w    int
}

type Buffer struct {
	head        *node
	tail        *node
	size        int
	minNodeSize int
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
