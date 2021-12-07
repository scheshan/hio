package poll

var DefaultWaitMs int64 = 30000

type PollerEvent struct {
	id  int
	typ int
}

func (t PollerEvent) Id() int {
	return t.id
}

func (t PollerEvent) CanRead() bool {
	return t.typ&1 > 0
}

func (t PollerEvent) CanWrite() bool {
	return t.typ&2 > 0
}
