package hio

import (
	"golang.org/x/sys/unix"
	"hio/buf"
	"sync"
)

type Conn struct {
	id       uint64
	sa       unix.Sockaddr
	fd       int
	flushing int32
	out      *buf.Buffer
	flush    *buf.Buffer
	loop     *EventLoop
	mutex    *sync.Mutex
	state    int
	attr     map[string]interface{}
}

func (t *Conn) Write(buffer *buf.Buffer) {
	t.mutex.Lock()
	t.out.Append(buffer)

	if t.flushing == 0 {
		t.flushing = 1
		t.loop.AddEvent(newWriteConnHandler(t))
	}

	t.mutex.Unlock()
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
	conn.attr = make(map[string]interface{})

	return conn
}
