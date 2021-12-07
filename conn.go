package hio

import (
	"golang.org/x/sys/unix"
	"hio/buf"
	"sync"
	"sync/atomic"
)

type Conn struct {
	id        uint64
	sa        unix.Sockaddr
	fd        int
	writeFlag int32 // 0 normal, 1 in EventLoop's write queue, 2 listen for write events
	out       *buf.Buffer
	flush     *buf.Buffer
	loop      *EventLoop
	mutex     *sync.Mutex
	state     int
	attr      map[string]interface{}
}

func (t *Conn) Write(buffer *buf.Buffer) {
	t.mutex.Lock()
	t.out.Append(buffer)
	t.mutex.Unlock()

	if atomic.CompareAndSwapInt32(&t.writeFlag, 0, 1) {
		t.loop.poll.AddReadWrite(t.fd)
	}
}

func (t *Conn) EventLoop() *EventLoop {
	return t.loop
}

func (t *Conn) release() {
	unix.Close(t.fd)

	if t.flush != nil {
		t.flush.Release()
	}
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
	conn.flush = buf.NewBuffer()
	conn.attr = make(map[string]interface{})

	return conn
}
