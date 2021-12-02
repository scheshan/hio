package hio

type bufferNode struct {
	next   *bufferNode
	origin *bufferNode
	b      []byte
	ref    int
	r      int
	w      int
}

func (t *bufferNode) readableBytes() int {
	return t.w - t.r
}

func (t *bufferNode) writableBytes() int {
	return len(t.b) - t.w
}

func (t *bufferNode) readBytes(n int) []byte {
	r := t.r
	t.r += n
	return t.b[r:t.r]
}

func (t *bufferNode) writeBytes(data []byte) int {
	n := copy(t.b[t.w:], data)
	t.w += n

	return n
}

func (t *bufferNode) writeByte(b ...byte) {
	for _, n := range b {
		t.b[t.w] = n
		t.w++
	}
}

func (t *bufferNode) release() {
	if t.ref <= 0 {
		return
	}

	t.ref--
	if t.ref == 0 {
		if t.origin != nil {
			t.origin.release()
		} else {
			pool.putBytes(t.b)
		}
		t.r = 0
		t.w = 0
		t.b = nil
		t.origin = nil
		t.next = nil
	}
}
