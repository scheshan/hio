package hio

type Buffer struct {
	b *Bytes
}

func NewBuffer() *Buffer {
	buf := new(Buffer)
	buf.b, _ = defaultBytesPool.Get(2 << 12)

	return buf
}
