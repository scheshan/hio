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

func (t *Conn) doClose() {
	syscall.Close(t.fd)
	t.in.Close()
	t.out.Close()
}

func newConn(id uint64, sa syscall.Sockaddr, fd int) *Conn {
	conn := &Conn{
		id:  id,
		sa:  sa,
		fd:  fd,
		in:  NewBuffer(),
		out: NewBuffer(),
	}

	return conn
}
