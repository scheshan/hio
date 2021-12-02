package hio

type bufferNode struct {
	next   *bufferNode
	origin *bufferNode
	b      []byte
	ref    int
	r      int
	w      int
}

func (t *bufferNode) copy(data []byte) int {
	n := copy(t.b[t.w:], data)
	t.w += n

	return n
}

func (t *bufferNode) readableBytes() int {
	return t.w - t.r
}

func (t *bufferNode) writableBytes() int {
	return len(t.b) - t.w
}

func (t *bufferNode) nextByte() byte {
	r := t.r
	t.r++
	b := t.b[r]

	return b
}

func (t *bufferNode) nextBytes(n int) []byte {
	r := t.r
	t.r += n
	b := t.b[r : r+n]

	return b
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
