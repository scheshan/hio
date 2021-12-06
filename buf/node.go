package buf

type node struct {
	b    []byte
	next *node
	r    int
	w    int
}

func (t *node) readableBytes() int {
	return t.w - t.r
}

func (t *node) writableBytes() int {
	return len(t.b) - t.w
}

func (t *node) readBytes(n int) []byte {
	r := t.r
	t.r += n
	return t.b[r:t.r]
}

func (t *node) writeBytes(data []byte) int {
	n := copy(t.b[t.w:], data)
	t.w += n

	return n
}

func (t *node) writeByte(b ...byte) {
	for _, n := range b {
		t.b[t.w] = n
		t.w++
	}
}

func (t *node) release() {
	returnBytes(t.b)
	t.b = nil
	t.r = 0
	t.w = 0
	t.next = nil

	returnNode(t)
}
