package hio

type Network interface {
	AddEvents(conns []*Conn) error
	RemoveEvents(conns []*Conn) error
	Wait(timeMs int) (fds []int, n int, err error)
}
