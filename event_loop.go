package hio

import (
	"sync/atomic"
	"syscall"
)

type EventLoop struct {
	nw      *network
	id      uint64
	connMap map[int]*Conn
	buf     []byte
	running int32
}

func (t *EventLoop) addConn(conn *Conn) {
	err := t.nw.addReadWrite(conn.fd)
	if err != nil {
		conn.doClose()
		return
	}

	t.connMap[conn.fd] = conn
}

func (t *EventLoop) deleteConn(conn *Conn) {
	delete(t.connMap, conn.fd)
	conn.doClose()
}

func (t *EventLoop) loop() {
	for t.running == 1 {
		events, err := t.nw.wait(networkWaitMs)
		if err != nil && err != syscall.EAGAIN {
			panic(err)
		}

		if len(events) == 0 {
			continue
		}

		for _, ev := range events {
			if ev.canRead() {
				t.readConn(ev.fd)
			}
		}
	}
}

func (t *EventLoop) readConn(fd int) {
	conn := t.connMap[fd]

	for {
		n, err := syscall.Read(fd, t.buf)
		if err != nil {
			if err != syscall.EAGAIN {
				t.deleteConn(conn)
			}
			return
		}

		conn.in.Write(t.buf[:n])
	}
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

func newEventLoop(id uint64) (*EventLoop, error) {
	nw, err := newNetwork()
	if err != nil {
		return nil, err
	}

	loop := &EventLoop{}
	loop.id = id
	loop.connMap = make(map[int]*Conn)
	loop.buf = make([]byte, 40960)
	loop.nw = nw

	return loop, nil
}
