package hio

import (
	"fmt"
	"sync"
	"syscall"
)

type Conn struct {
	id        uint64
	sa        syscall.Sockaddr
	fd        int
	out       *Buffer
	flush     *Buffer
	loop      *EventLoop
	flushing  bool
	flushMask bool
	mutex     *sync.Mutex
}

func (t *Conn) Close() error {
	return nil
}

func (t *Conn) EventLoop() *EventLoop {
	return t.loop
}

func (t *Conn) Write(buf *Buffer) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.Append(buf)
}

func (t *Conn) WriteAndFlush(buf *Buffer) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.out.Append(buf)

	return t.doFlush()
}

func (t *Conn) Flush() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.doFlush()
}

func (t *Conn) String() string {
	return fmt.Sprintf("%v", t.id)
}

func (t *Conn) doClose() {
	syscall.Close(t.fd)
	t.out.Release()
}

func (t *Conn) doFlush() error {
	if t.flushing {
		return nil
	}
	t.flushing = true

	t.flush.Append(t.out)

	return t.flushToFile()
}

func (t *Conn) flushToFile() error {
	err := t.flush.copyToFile(t.fd)
	if err != nil {
		if err == syscall.EAGAIN {
			t.loop.markWrite(t, true)
			return nil
		}

		t.loop.onConnError(t, err)
		return err
	}

	if t.flush.ReadableBytes() == 0 {
		t.flushing = false
	}

	t.loop.markWrite(t, false)
	return nil
}

func newConn(id uint64, sa syscall.Sockaddr, fd int) *Conn {
	conn := &Conn{
		id:    id,
		sa:    sa,
		fd:    fd,
		out:   &Buffer{},
		flush: &Buffer{},
		mutex: &sync.Mutex{},
	}

	return conn
}
