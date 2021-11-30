package hio

type Buffer struct {
	b *Bytes
	r int
	w int
}

func (t *Buffer) Close() {
	if t.b != nil {
		t.b.Release()
		t.b = nil
	}
}

func (t *Buffer) Write(data []byte) error {
	if err := t.checkSize(len(data)); err != nil {
		return err
	}

	copy(t.b.b[t.w:], data)
	t.w += len(data)

	return nil
}

func (t *Buffer) checkSize(size int) error {
	if len(t.b.b)-t.w <= size {
		return nil
	}

	if t.w-t.r <= size {
		copy(t.b.b, t.b.b[t.r:t.w])
		t.w -= t.r
		t.r = 0
		return nil
	}

	newSize := t.w - t.r + size
	b, err := defaultBytesPool.Get(newSize)
	if err != nil {
		return err
	}

	copy(b.b, t.b.b[t.r:t.w])
	t.w -= t.r
	t.r = 0
	t.b.Release()
	t.b = b

	return nil
}

func NewBuffer() *Buffer {
	buf := new(Buffer)
	buf.b, _ = defaultBytesPool.Get(2 << 12)

	return buf
}
