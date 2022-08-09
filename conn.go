package hio

import (
	"fmt"
	"golang.org/x/sys/unix"
	"sync/atomic"
)

var connId uint64

type Conn interface {
	fmt.Stringer
	Id() uint64
}

type conn struct {
	id   uint64
	sa   unix.Sockaddr
	fd   int
	loop *eventLoop
}

func (t *conn) String() string {
	return ""
}

func (t *conn) Id() uint64 {
	return t.id
}

func newConn(fd int, sa unix.Sockaddr) *conn {
	c := &conn{
		id: atomic.AddUint64(&connId, 1),
		sa: sa,
		fd: fd,
	}
	return c
}
