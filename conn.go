package hio

import (
	"errors"
	"golang.org/x/sys/unix"
	"hio/buf"
	"sync"
	"sync/atomic"
)

var ErrConnNonActive = errors.New("failed to operate on a non-active conn")

type Conn struct {
	id        uint64
	sa        unix.Sockaddr
	fd        int
	writeFlag int32 // 0 normal, 1 listen for write events
	out       *buf.Buffer
	loop      *EventLoop
	mutex     *sync.Mutex
	state     int32 // 1 opened, 0 half closed, -1 error, -2 closed
	attr      map[string]interface{}
}

func (t *Conn) Write(buffer *buf.Buffer) error {
	if !t.Active() {
		return ErrConnNonActive
	}

	if buffer.ReadableBytes() > 0 {
		b := buf.NewBuffer()
		b.Append(buffer)
		t.loop.QueueEvent(func() {
			t.out.Append(b)
			b.Release()
			t.loop.writeConn(t)
		})
	}

	return nil
}

func (t *Conn) EventLoop() *EventLoop {
	return t.loop
}

func (t *Conn) Active() bool {
	return t.state > 0
}

func (t *Conn) Close() error {
	if !atomic.CompareAndSwapInt32(&t.state, 1, 0) {
		return ErrConnNonActive
	}

	t.loop.QueueEvent(connEventFunc(t.loop.closeConn, t))
	return nil
}

func (t *Conn) release() {
	unix.Close(t.fd)

	if t.out != nil {
		t.out.Release()
	}
}

func newConn(id uint64, fd int, sa unix.Sockaddr) *Conn {
	conn := &Conn{}
	conn.id = id
	conn.fd = fd
	conn.sa = sa
	conn.mutex = &sync.Mutex{}

	conn.out = buf.NewBuffer()
	conn.attr = make(map[string]interface{})

	return conn
}
