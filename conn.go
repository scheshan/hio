package hio

import "syscall"

type Conn struct {
	id   uint64
	sa   syscall.Sockaddr
	fd   int
	in   *Buffer
	out  *Buffer
	loop *EventLoop
}

func (t *Conn) Close() error {
	return nil
}

func (t *Conn) EventLoop() *EventLoop {
	return t.loop
}

func (t *Conn) doClose() {
	syscall.Close(t.fd)
	t.in.Release()
	t.out.Release()
}

func newConn(id uint64, sa syscall.Sockaddr, fd int) *Conn {
	conn := &Conn{
		id:  id,
		sa:  sa,
		fd:  fd,
		in:  &Buffer{},
		out: &Buffer{},
	}

	return conn
}
