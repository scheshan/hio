package hio

import (
	"golang.org/x/sys/unix"
	"hio/buf"
	"log"
	"sync"
)

type EventHandler interface {
	Handle()
}

type readConnHandler struct {
	loop *EventLoop
	conn *Conn
}

var readConnHandlerPool = &sync.Pool{
	New: func() interface{} {
		return &readConnHandler{}
	},
}

func (t *readConnHandler) Handle() {
	defer t.release()
	b := buf.NewBuffer()
	defer b.Release()

	for i := 0; i < 8; i++ {
		_, err := b.WriteFromFile(t.conn.fd)
		if err != nil {
			if err == unix.EAGAIN {
				break
			}

			log.Printf("read conn error: %v", err)
			t.loop.deleteConn(t.conn)
			return
		}
	}

	if t.loop.opt.OnSessionRead != nil {
		t.loop.opt.OnSessionRead(t.conn, b)
	}
}

func (t *readConnHandler) release() {
	t.loop = nil
	t.conn = nil
	readConnHandlerPool.Put(t)
}

type writeConnHandler struct {
	loop *EventLoop
	conn *Conn
}

var writeConnHandlerPool = &sync.Pool{
	New: func() interface{} {
		return &writeConnHandler{}
	},
}

func (t *writeConnHandler) Handle() {
	defer t.release()

	conn := t.conn

	for i := 0; i < 8; i++ {
		if conn.out.ReadableBytes() > 0 {
			conn.mutex.Lock()
			conn.flush.Append(conn.out)
			conn.mutex.Unlock()
		}

		if conn.flush.ReadableBytes() == 0 {
			return
		}

		_, err := conn.flush.ReadToFile(conn.fd)
		if err != nil {
			if err == unix.EAGAIN {
				break
			}

			log.Printf("write conn error: %v", err)
			t.loop.deleteConn(t.conn)
			return
		}
	}
}

func (t *writeConnHandler) release() {
	t.loop = nil
	t.conn = nil
	writeConnHandlerPool.Put(t)
}

func newReadConnHandler(conn *Conn) EventHandler {
	h := readConnHandlerPool.Get().(*readConnHandler)
	h.conn = conn
	h.loop = conn.loop

	return h
}

func newWriteConnHandler(conn *Conn) EventHandler {
	h := writeConnHandlerPool.Get().(*writeConnHandler)
	h.conn = conn
	h.loop = conn.loop

	return h
}
