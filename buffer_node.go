package hio

type bufferNode struct {
	next  *bufferNode
	bytes *Bytes
	ref   int
	r     int
	w     int
}

func (t *bufferNode) copy(data []byte) int {
	n := copy(t.bytes.SliceStart(t.w), data)
	t.w += n

	return n
}

func (t *bufferNode) readableBytes() int {
	return t.w - t.r
}

func (t *bufferNode) writableBytes() int {
	return t.bytes.Len() - t.w
}

func (t *bufferNode) nextByte() byte {
	r := t.r
	t.r++
	b := t.bytes.Index(r)

	return b
}

func (t *bufferNode) nextBytes(n int) []byte {
	r := t.r
	t.r += n
	b := t.bytes.Slice(r, r+n)

	return b
}
