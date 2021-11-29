package hio

var networkWaitMs int64 = 30_000

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
