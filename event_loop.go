package hio

import (
	"container/list"
	"golang.org/x/sys/unix"
	"hio/poll"
	"sync"
	"sync/atomic"
)

type EventLoop struct {
	poll    *poll.Poller
	id      uint64
	connMap map[int]*Conn
	running int32
	opt     ServerOptions
	events  []*list.List
	awake   int32
	mutex   *sync.Mutex
}

func (t *EventLoop) Id() uint64 {
	return t.id
}

func (t *EventLoop) AddEvent(h EventHandler) {
	t.mutex.Lock()

	t.addEvent(1, h)

	t.mutex.Unlock()

	if atomic.CompareAndSwapInt32(&t.awake, 0, 1) {
		t.poll.Wakeup()
	}
}

func (t *EventLoop) addEvent(ind int, h EventHandler) {
	if t.events[ind] == nil {
		t.events[ind] = &list.List{}
	}
	t.events[ind].PushBack(h)
}

func (t *EventLoop) bindConn(conn *Conn) error {
	if err := t.poll.AddRead(conn.fd); err != nil {
		return err
	}

	conn.loop = t
	t.connMap[conn.fd] = conn

	if t.opt.OnSessionCreated != nil {
		t.opt.OnSessionCreated(conn)
	}

	return nil
}

func (t *EventLoop) deleteConn(conn *Conn) {
	t.poll.RemoveReadWrite(conn.fd)

	delete(t.connMap, conn.fd)
	conn.release()

	if t.opt.OnSessionClosed != nil {
		t.opt.OnSessionClosed(conn)
	}
}

func (t *EventLoop) loop() {
	for t.running == 1 {
		t.addIOEvents()
		t.addCustomEvents()

		t.processEvents()
	}

	t.release()
}

func (t *EventLoop) addIOEvents() {
	events, err := t.poll.Wait(networkWaitMs)
	if err != nil && err != unix.EAGAIN && err != unix.EINTR {
		return
	}

	for _, event := range events {
		conn := t.connMap[event.Id()]
		if conn == nil {
			continue
		}

		if event.CanRead() {
			t.addEvent(0, newReadConnHandler(conn))
		}
		if event.CanWrite() {
			t.addEvent(0, newWriteConnHandler(conn))
		}
	}
}

func (t *EventLoop) addCustomEvents() {
	t.mutex.Lock()
	list := t.events[1]
	t.events[1] = nil
	t.mutex.Unlock()

	for list != nil && list.Front() != nil {
		f := list.Front()
		t.addEvent(0, f.Value.(EventHandler))
		list.Remove(f)
	}
}

func (t *EventLoop) processEvents() {
	list := t.events[0]

	for list != nil && list.Front() != nil {
		f := list.Front()
		h := f.Value.(EventHandler)

		h.Handle()

		list.Remove(f)
	}
}

func (t *EventLoop) shutdown() {
	t.running = 0
}

func (t *EventLoop) release() {
	t.poll.Close()
	for _, conn := range t.connMap {
		conn.release()
	}

	t.connMap = nil
}

func newEventLoop(id uint64, opt ServerOptions) (*EventLoop, error) {
	poll, err := poll.NewPoller()
	if err != nil {
		return nil, err
	}

	loop := &EventLoop{}
	loop.id = id
	loop.connMap = make(map[int]*Conn)
	loop.poll = poll
	loop.opt = opt
	loop.events = make([]*list.List, 2)
	loop.mutex = &sync.Mutex{}

	return loop, nil
}
