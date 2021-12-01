package hio

type Bytes struct {
	origin []byte
	buf    []byte
	ind    int
	ref    int
}

func (t *Bytes) incrRef() {
	t.ref++
}

func (t *Bytes) decrRef() {
	if t.ref <= 0 {
		return
	}

	t.ref--
}

func (t *Bytes) Release() {
	t.decrRef()

	if t.ref == 0 {
		defaultBytesPool.put(t)
	}
}

func (t *Bytes) Data() []byte {
	return t.buf
}

func (t *Bytes) Slice(start, end int) []byte {
	return t.buf[start:end]
}

func (t *Bytes) SliceStart(start int) []byte {
	return t.buf[start:]
}

func (t *Bytes) SliceEnd(end int) []byte {
	return t.buf[:end]
}

func (t *Bytes) Len() int {
	return len(t.buf)
}

func (t *Bytes) Index(ind int) byte {
	return t.buf[ind]
}

func (t *Bytes) CopyFrom(data []byte) int {
	return copy(t.buf, data)
}

func (t *Bytes) CopyStartFrom(start int, data []byte) int {
	return copy(t.buf[start:], data)
}

func (t *Bytes) init(size int) {
	t.buf = t.origin[:size]
}

func Malloc(size int) *Bytes {
	b := defaultBytesPool.get(size)
	b.incrRef()

	return b
}
