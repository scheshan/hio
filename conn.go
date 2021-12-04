package hio

import (
	"errors"
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
	state     int
}

//TODO close connection gracefully
func (t *Conn) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if !t.Active() {
		return errors.New("cannot close a non-active connection")
	}
	t.state = -2

	t.mutex.Unlock()

	if t.flushing {
		return nil
	}
	if t.out.ReadableBytes() > 0 {
		return t.Flush()
	} else {
		t.flush.Append(t.out)
		return t.doFlush()
	}
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
	t.flush.Append(t.out)

	return t.doFlush()
}

func (t *Conn) Flush() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.flush.Append(t.out)

	return t.doFlush()
}

func (t *Conn) String() string {
	return fmt.Sprintf("%v", t.id)
}

func (t *Conn) Active() bool {
	return t.state >= 0
}

func (t *Conn) doClose() {
	t.out.Release()
	t.flush.Release()
	syscall.Close(t.fd)
}

func (t *Conn) doFlush() error {
	if t.flushing {
		return nil
	}
	t.flushing = true

	return t.loop.flushConn(t)
}

func (t *Conn) flushCompleted() {
	t.flushing = false
	t.loop.markWrite(t, t.flush.ReadableBytes() > 0)
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
