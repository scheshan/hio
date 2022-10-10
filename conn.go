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

	return ""
}

func (t *conn) Id() uint64 {
	return t.id
}

func (t *conn) Close() {

}

func newConn(fd int, sa unix.Sockaddr) *conn {
	c := &conn{
		id: atomic.AddUint64(&connId, 1),
		sa: sa,
		fd: fd,
		out: buffer.NewWithOptions(buffer.Options{
			MinAllocSize: 4096,
			MaxSize:      40960000,
		}),
	}
	return c
}
