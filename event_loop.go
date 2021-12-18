package hio

import (
	"container/list"
	"github.com/scheshan/poll"
	"golang.org/x/sys/unix"
	"log"
	"sync"
)

type EventLoop struct {
	poller  *poll.Poller
	id      uint64
	connMap map[int]*Conn
	running int32
	handler EventHandler
	events  *list.List
	awake   int32
	mutex   *sync.Mutex
	buf     []byte
}

func (t *EventLoop) Id() uint64 {
	return t.id
}

//func (t *EventLoop) QueueEvent(f EventFunc) {
//	awake := false
//
//	t.mutex.Lock()
//	if t.events == nil {
//		t.events = &list.List{}
//	}
//	t.events.PushBack(f)
//
//	if t.awake == 0 {
//		t.awake = 1
//		awake = true
//	}
//	t.mutex.Unlock()
//
//	if awake {
//		t.poll.Wakeup()
//	}
//}

func (t *EventLoop) bindConn(conn *Conn) error {
	if err := t.poller.Add(conn.fd); err != nil {
		return err
	}

	conn.state = 1
	conn.loop = t
	t.connMap[conn.fd] = conn

	if t.handler.SessionCreated != nil {
		t.handler.SessionCreated(conn)
	}

	return nil
}

func (t *EventLoop) deleteConn(conn *Conn) {
	t.poller.Delete(conn.fd)

	delete(t.connMap, conn.fd)
	conn.release()

	if t.handler.SessionClosed != nil {
		t.handler.SessionClosed(conn)
	}
}

func (t *EventLoop) loop() {
	for t.running == 1 {
		t.processIOEvents()
		//t.processUserEvents()
	}

	t.release()
}

func (t *EventLoop) processIOEvents() {
	t.poller.Wait(5000, func(fd int, flag poll.Flag) error {
		conn := t.connMap[fd]
		if conn == nil {
			return nil
		}

		if flag.CanRead() {
			t.handleConnRead(conn)
		}
		if flag.CanWrite() {
			t.handleConnWrite(conn)
		}

		return nil
	})
}

//func (t *EventLoop) processUserEvents() {
//	t.mutex.Lock()
//	list := t.events
//	t.events = nil
//	t.awake = 0
//	t.mutex.Unlock()
//
//	for list != nil && list.Front() != nil {
//		f := list.Front()
//		ef := f.Value.(EventFunc)
//		ef()
//		list.Remove(f)
//	}
//
//}

func (t *EventLoop) handleConnRead(conn *Conn) {
	n, err := unix.Read(conn.fd, t.buf)
	if err != nil {
		if err == unix.EAGAIN {
			return
		}

		log.Printf("read conn[%v] error: %v", conn, err)
		conn.state = -1
		t.deleteConn(conn)
		return
	}
	if n == 0 {
		t.handleConnClose(conn)
		return
	}

	if t.handler.SessionRead != nil {
		data := t.handler.SessionRead(conn, t.buf[:n])
		if data != nil {
			t.writeConn(conn, data)
		}
	}
}

func (t *EventLoop) handleConnWrite(conn *Conn) {
	if !conn.Active() {
		return
	}

	t.writeConn0(conn)

	if conn.out.Len() == 0 {
		conn.writeFlag = 0
		t.poller.DisableWrite(conn.fd)
	}
}

func (t *EventLoop) handleConnClose(conn *Conn) {
	conn.state = 0
	t.deleteConn(conn)
}

func (t *EventLoop) writeConn(conn *Conn, data []byte) {
	if !conn.Active() {
		return
	}

	if conn.out.Len() == 0 && conn.writeFlag == 0 {
		n, err := unix.Write(conn.fd, data)
		if err != nil && err != unix.EAGAIN {
			log.Printf("write conn[%v] error: %v", conn, err)
			t.closeConn(conn)
			return
		}
		if n == 0 {
			t.handleConnClose(conn)
			return
		}

		if n < len(data) {
			conn.out.WriteBytes(data[n:])
		}
	}

	if conn.out.Len() > 0 {
		conn.writeFlag = 1
		t.poller.EnableWrite(conn.fd)
	}
}

func (t *EventLoop) writeConn0(conn *Conn) {
	n, err := conn.out.CopyToFile(conn.fd)
	if err != nil {
		if err != unix.EAGAIN {
			log.Printf("write conn[%v] error: %v", conn, err)
			conn.state = -1
			t.deleteConn(conn)
		}
		return
	}

	if n == 0 {
		t.deleteConn(conn)
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
	t.poller.Close()
	for _, conn := range t.connMap {
		conn.release()
	}

	t.connMap = nil
}

func newEventLoop(id uint64, handler EventHandler) (*EventLoop, error) {
	poller, err := poll.NewPoller()
	if err != nil {
		return nil, err
	}

	loop := &EventLoop{}
	loop.id = id
	loop.connMap = make(map[int]*Conn)
	loop.poller = poller
	loop.handler = handler
	loop.mutex = &sync.Mutex{}
	loop.running = 1
	loop.buf = make([]byte, 81920)

	return loop, nil
}
