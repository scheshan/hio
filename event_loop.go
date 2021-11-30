package hio

import (
	"sync/atomic"
	"syscall"
)

type EventLoop struct {
	nw      *network
	id      uint64
	connMap map[int]*Conn
	buf     *Bytes
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

func (t *EventLoop) loop() {
	for t.running == 1 {

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
			syscall.Close(conn.fd)
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
	loop.buf, _ = defaultBytesPool.Get(4096)
	loop.nw = nw

	return loop, nil
}
