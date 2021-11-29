package hio

type networkEvent struct {
	fd int
	ev int
}

func (t *networkEvent) canRead() bool {
	return t.ev|1 == 1
}

func (t *networkEvent) canWrite() bool {
	return t.ev|2 == 2
}
