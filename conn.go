package hio

import (
	"fmt"
	"syscall"
)

type ConnState int

const (
	ConnState_Open ConnState = iota
	ConnState_Close
)

type Conn struct {
	id    int
	loop  *EventLoop
	fd    int
	addr  syscall.Sockaddr
	in    []byte
	out   []byte
	state ConnState
}

func (t *Conn) Write(data []byte) {
	t.out = append(t.out, data...)
}

func (t *Conn) Close() {
	t.state = ConnState_Close
	t.loop.tryCloseConn(t)
}

func (t *Conn) String() string {
	return fmt.Sprintf("%v", t.id)
}

func newConn(loop *EventLoop, id int, fd int, addr syscall.Sockaddr) *Conn {
	conn := &Conn{
		id:   id,
		loop: loop,
		fd:   fd,
		addr: addr,
		in:   make([]byte, 4096, 4096),
		out:  make([]byte, 0, 4096),
	}

	return conn
}
