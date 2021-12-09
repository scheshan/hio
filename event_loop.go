package hio

import (
	"container/list"
	"github.com/scheshan/hio/buf"
	"github.com/scheshan/hio/poll"
	"golang.org/x/sys/unix"
	"log"
	"sync"
)

type EventLoop struct {
	poll    *poll.Poller
	id      uint64
	connMap map[int]*Conn
	running int32
	opt     ServerOptions
	events  *list.List
	awake   int32
	mutex   *sync.Mutex
}

func (t *EventLoop) Id() uint64 {
	return t.id
}

func (t *EventLoop) QueueEvent(f EventFunc) {
	awake := false

	t.mutex.Lock()
	if t.events == nil {
		t.events = &list.List{}
	}
	t.events.PushBack(f)

	if t.awake == 0 {
		t.awake = 1
		awake = true
	}
	t.mutex.Unlock()

	if awake {
		t.poll.Wakeup()
	}
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
			t.handleConnRead(conn)
		}
		if event.CanWrite() {
			t.handleConnWrite(conn)
		}
	}
}

func (t *EventLoop) processUserEvents() {
	t.mutex.Lock()
	list := t.events
	t.events = nil
	t.awake = 0
	t.mutex.Unlock()

	for list != nil && list.Front() != nil {
		f := list.Front()
		ef := f.Value.(EventFunc)
		ef()
		list.Remove(f)
	}

}

func (t *EventLoop) handleConnRead(conn *Conn) {
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

func (t *EventLoop) handleConnWrite(conn *Conn) {
	if !conn.Active() {
		return
	}

	t.writeConn0(conn)

	if conn.out.ReadableBytes() == 0 {
		conn.writeFlag = 0
		t.poll.DisableWrite(conn.fd)
	}
}

func (t *EventLoop) writeConn(conn *Conn, buffer *buf.Buffer) {
	if !conn.Active() {
		return
	}

	conn.out.Append(buffer)
	buffer.Release()

	if conn.out.ReadableBytes() == 0 || conn.writeFlag == 1 {
		return
	}

	t.writeConn0(conn)

	if conn.out.ReadableBytes() > 0 {
		conn.writeFlag = 1
		t.poll.EnableWrite(conn.fd)
	}
}

func (t *EventLoop) writeConn0(conn *Conn) {
	for i := 0; i < 8; i++ {
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

		if conn.out.ReadableBytes() == 0 {
			break
		}
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
	loop.mutex = &sync.Mutex{}
	loop.running = 1

	return loop, nil
}
