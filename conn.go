package hio

import "syscall"

type Conn struct {
	loop *EventLoop
	fd   int
	addr syscall.Sockaddr
	in   []byte
	out  []byte
}

func (t *Conn) Write(data []byte) {
	t.out = append(t.out, data...)
}

func newConn(loop *EventLoop, fd int, addr syscall.Sockaddr) *Conn {
	conn := &Conn{
		loop: loop,
		fd:   fd,
		addr: addr,
		in:   make([]byte, 4096, 4096),
		out:  make([]byte, 0, 4096),
	}

	return conn
}
