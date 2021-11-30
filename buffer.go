package hio

type Buffer struct {
	b *Bytes
}

func (t *Buffer) Close() {
	if t.b != nil {
		t.b.Release()
		t.b = nil
	}
}

func NewBuffer() *Buffer {
	buf := new(Buffer)
	buf.b, _ = defaultBytesPool.Get(2 << 12)

	return buf
}
