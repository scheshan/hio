package hio

import (
	"golang.org/x/sys/unix"
	"hio/buf"
	"sync"
)

type Conn struct {
	id        uint64
	sa        unix.Sockaddr
	fd        int
	writeFlag int32 // 0 normal, 1 listen for write events
	out       *buf.Buffer
	loop      *EventLoop
	mutex     *sync.Mutex
	state     int // 1 opened, 0 closed, -1 error
	attr      map[string]interface{}
}

func (t *Conn) Write(buffer *buf.Buffer) {
	if buffer.ReadableBytes() > 0 {
		b := buf.NewBuffer()
		b.Append(buffer)
		t.loop.QueueEvent(func() {
			t.out.Append(b)
			b.Release()
			t.loop.writeConn(t)
		})
	}
}

func (t *Conn) EventLoop() *EventLoop {
	return t.loop
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
