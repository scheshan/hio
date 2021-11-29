package hio

type Bytes struct {
	b   []byte
	ind int
	p   *BytesPool
}

func (t *Bytes) Release() {
	t.p.put(t)
}
