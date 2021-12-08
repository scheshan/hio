package hio

import (
	"container/list"
	"golang.org/x/sys/unix"
	"hio/buf"
	"hio/poll"
	"log"
	"sync"
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

func (t *EventLoop) QueueEvent(f EventFunc) {
	awake := false

	t.mutex.Lock()
	t.queueEvent(1, f)
	if t.awake == 0 {
		t.awake = 1
		awake = true
	}
	t.mutex.Unlock()

	if awake {
		t.poll.Wakeup()
	}
}

func (t *EventLoop) queueEvent(ind int, f EventFunc) {
	if t.events[ind] == nil {
		t.events[ind] = &list.List{}
	}
	t.events[ind].PushBack(f)
}

func (t *EventLoop) bindConn(conn *Conn) error {
	if err := t.poll.Add(conn.fd); err != nil {
		return err
	}

	conn.state = 1
	conn.loop = t
	t.connMap[conn.fd] = conn

	if t.opt.OnSessionCreated != nil {
		t.opt.OnSessionCreated(conn)
	}

	return nil
}

func (t *EventLoop) deleteConn(conn *Conn) {
	t.poll.Delete(conn.fd)

	delete(t.connMap, conn.fd)
	conn.release()

	if t.opt.OnSessionClosed != nil {
		t.opt.OnSessionClosed(conn)
	}
}

func (t *EventLoop) loop() {
	for t.running == 1 {
		t.processIOEvents()
		t.processUserEvents()

		t.processEvents()
	}

	t.release()
}

func (t *EventLoop) processIOEvents() {
	events, err := t.poll.Wait(poll.DefaultWaitMs)
	if err != nil && err != unix.EAGAIN && err != unix.EINTR {
		return
	}

	for _, event := range events {
		conn := t.connMap[event.Id()]
		if conn == nil {
			continue
		}

		if event.CanRead() {
			t.queueEvent(0, connEventFunc(t.readConn, conn))
		}
		if event.CanWrite() {
			t.queueEvent(0, connEventFunc(t.writeConn, conn))
		}
	}
}

func (t *EventLoop) processUserEvents() {
	t.mutex.Lock()
	list := t.events[1]
	t.events[1] = nil
	t.awake = 0
	t.mutex.Unlock()

	for list != nil && list.Front() != nil {
		f := list.Front()
		t.queueEvent(0, f.Value.(EventFunc))
		list.Remove(f)
	}

}

func (t *EventLoop) processEvents() {
	list := t.events[0]

	for list != nil && list.Front() != nil {
		f := list.Front()
		h := f.Value.(EventFunc)

		h()

		list.Remove(f)
	}
}

func (t *EventLoop) readConn(conn *Conn) {
	b := buf.NewBuffer()
	defer b.Release()

	for i := 0; i < 8; i++ {
		n, err := b.WriteFromFile(conn.fd)
		if err != nil {
			if err == unix.EAGAIN {
				break
			}

			log.Printf("read conn[%v] error: %v", conn, err)
			conn.state = -1
			t.deleteConn(conn)
			return
		}
		if n == 0 {
			//close by foreign host
			conn.state = 0
			t.deleteConn(conn)
			return
		}
	}

	if t.opt.OnSessionRead != nil {
		t.opt.OnSessionRead(conn, b)
	}
}

func (t *EventLoop) writeConn(conn *Conn) {
	if conn.writeFlag == 0 {
		conn.writeFlag = 1
		err := t.poll.EnableWrite(conn.fd)
		if err != nil {
			log.Printf("add write failed: %v", err)
		}
		return
	}

	for i := 0; i < 8; i++ {
		if conn.out.ReadableBytes() == 0 {
			break
		}

		_, err := conn.out.ReadToFile(conn.fd)
		if err != nil {
			if err == unix.EAGAIN {
				break
			}

			log.Printf("write conn[%v] error: %v", conn, err)
			conn.state = -1
			t.deleteConn(conn)
			return
		}
		//TODO handle write to a closed fd
	}

	if conn.out.ReadableBytes() == 0 {
		err := t.poll.DisableWrite(conn.fd)
		if err != nil {
			log.Printf("remove write failed: %v", err)
		}
		conn.writeFlag = 0
	}
}

func (t *EventLoop) closeConn(conn *Conn) {
	conn.state = -2
	t.deleteConn(conn)
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
	loop.running = 1

	return loop, nil
}
