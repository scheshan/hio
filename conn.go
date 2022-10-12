package hio

import (
	"fmt"
	"github.com/scheshan/buffer"
	"golang.org/x/sys/unix"
	"sync/atomic"
)

var connId uint64

type Conn interface {
	fmt.Stringer
	Id() uint64
	Close()
	Write(data []byte) error
}

type conn struct {
	id   uint64
	sa   unix.Sockaddr
	fd   int
	loop *eventLoop
	in   *buffer.Buffer
	out  *buffer.Buffer
}

func (t *conn) String() string {
	return fmt.Sprintf("conn-%v", t.id)
}

func (t *conn) Id() uint64 {
	return t.id
}

func (t *conn) Close() {
	t.loop.AddTask(func() error {
		t.loop.closeConn(t, nil)
		return nil
	})
}

func (t *conn) Write(data []byte) error {
	t.loop.AddTask(func() error {
		t.loop.writeConn(t, data)
		return nil
	})
	return nil
}

func newConn(fd int, sa unix.Sockaddr) *conn {
	c := &conn{
		id: atomic.AddUint64(&connId, 1),
		sa: sa,
		fd: fd,
	}
	return c
}
