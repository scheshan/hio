package hio

import (
	"log"
	"sync/atomic"
	"syscall"
)

type EventLoop struct {
	nw      *network
	id      uint64
	connMap map[int]*Conn
	buf     []byte
	running int32
	opt     ServerOptions
}

func (t *EventLoop) addConn(conn *Conn) {
	if err := syscall.SetNonblock(conn.fd, true); err != nil {
		conn.doClose()
		return
	}

	if err := t.nw.addRead(conn.fd); err != nil {
		conn.doClose()
		return
	}

	conn.loop = t
	t.connMap[conn.fd] = conn

	if t.opt.OnSessionCreated != nil {
		t.opt.OnSessionCreated(conn)
	}
}

func (t *EventLoop) deleteConn(conn *Conn) {
	delete(t.connMap, conn.fd)
	conn.doClose()
}

func (t *EventLoop) loop() {
	for t.running == 1 {
		events, err := t.nw.wait(networkWaitMs)
		if err != nil {
			if err == syscall.EAGAIN || err == syscall.EINTR {
				continue
			}
			if err == syscall.EBADF {
				return
			}

			panic(err)
		}

		if len(events) == 0 {
			continue
		}

		for _, ev := range events {
			if ev.canRead() {
				conn := t.connMap[ev.fd]
				if conn != nil {
					t.readConn(conn)
				}
			}
		}
	}
}

func (t *EventLoop) readConn(conn *Conn) {
	buf := pool.getBuffer()
	defer func() {
		buf.Release()
	}()

	for {
		n, err := syscall.Read(conn.fd, t.buf)
		if err != nil {
			if err == syscall.EAGAIN {
				break
			}
			t.onConnError(conn, err)
			return
		}
		if n == 0 {
			t.deleteConn(conn)
			return
		}

		if t.opt.OnSessionRead != nil {
			buf.WriteBytes(t.buf[:n])
		}
	}
	if t.opt.OnSessionRead != nil {
		t.opt.OnSessionRead(conn, buf)
	}
}

func (t *EventLoop) writeConn(conn *Conn) {
	conn.mutex.Lock()
	defer conn.mutex.Unlock()

	for {
		err := conn.flushToFile()
		if err != nil {
			if err == syscall.EAGAIN || err == syscall.EINTR {
				break
			}

			t.onConnError(conn, err)
			return
		}
	}

	t.markWrite(conn, conn.flush.ReadableBytes() > 0)
}

func (t *EventLoop) run() {
	if !atomic.CompareAndSwapInt32(&t.running, 0, 1) {
		return
	}

	go t.loop()
}

func (t *EventLoop) shutdown() {
	if !atomic.CompareAndSwapInt32(&t.running, 1, 0) {
		return
	}

	if t.nw != nil {
		t.nw.shutdown()
	}

	if t.connMap != nil {
		cm := t.connMap
		t.connMap = nil
		for _, conn := range cm {
			conn.doClose()
		}
	}
}

func (t *EventLoop) markWrite(conn *Conn, mask bool) {
	var err error
	flushMask := conn.flushMask
	if mask {
		if !flushMask {
			flushMask = true
			err = t.nw.addWrite(conn.fd)
		}
	} else {
		if flushMask {
			flushMask = false
			err = t.nw.removeWrite(conn.fd)
		}
	}
	conn.flushMask = flushMask
	if err != nil {
		t.onConnError(conn, err)
	}
}

func (t *EventLoop) onConnError(conn *Conn, err error) {
	log.Printf("error occours when operate conn[%s]: %v", conn, err)

	t.deleteConn(conn)
}

func (t *EventLoop) Id() uint64 {
	return t.id
}

func newEventLoop(id uint64, opt ServerOptions) (*EventLoop, error) {
	nw, err := newNetwork()
	if err != nil {
		return nil, err
	}

	loop := &EventLoop{}
	loop.id = id
	loop.connMap = make(map[int]*Conn)
	loop.buf = make([]byte, 40960)
	loop.nw = nw
	loop.opt = opt

	return loop, nil
}
