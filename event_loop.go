package hio

import (
	"fmt"
	"github.com/scheshan/poll"
	"golang.org/x/sys/unix"
	"log"
	"sync/atomic"
)

var eventLoopId uint64

type EventLoop interface {
}

type eventLoop struct {
	id      uint64
	poller  *poll.Poller
	connMap map[int]*conn
	state   int32
	handler EventHandler
	buf     []byte
	tasks   *taskQueue
	wakeup  int32
}

func (t *eventLoop) String() string {
	return fmt.Sprintf("EventLoop-%v", t.id)
}

func (t *eventLoop) Loop() {
	defer func() {
		t.poller.Close()
	}()

	for t.state == 0 {
		err := t.poller.Wait(30000, t.callback)
		switch err {
		case nil:
		case unix.EAGAIN:
		case unix.EINTR:
		default:
			return
		}

		t.handleTask()
	}
}

func (t *eventLoop) AddConn(conn *conn) {
	if err := t.poller.Add(conn.fd); err != nil {
		unix.Close(conn.fd)
		return
	}

	conn.loop = t
	t.connMap[conn.fd] = conn

	if t.handler.ConnCreate != nil {
		t.handler.ConnCreate(conn)
	}
}

func (t *eventLoop) Shutdown() {

}

func (t *eventLoop) AddTask(fn func() error) {
	t.tasks.Enqueue(fn)
	if atomic.CompareAndSwapInt32(&t.wakeup, 0, 1) {
		log.Printf("%s wakeup the poller", t)

		if err := t.poller.Wakeup(); err != nil {
			if err == unix.EAGAIN || err == unix.EINTR {
				return
			}
			log.Fatalf("wakeup eventloop failed: %v", err)
		}
	}
}

func (t *eventLoop) callback(fd int, flag poll.Flag) error {
	conn, ok := t.connMap[fd]
	if !ok {
		return nil
	}

	if flag.CanRead() {
		t.handleConnRead(conn)
	}
	if flag.CanWrite() {
		t.handleConnWrite(conn)
	}

	return nil
}

func (t *eventLoop) handleConnRead(conn *conn) {
	n, err := unix.Read(conn.fd, t.buf)
	if err != nil {
		if err != unix.EAGAIN && err != unix.EINTR {
			t.closeConn(conn)
			return
		}
	}

	if n > 0 {
		data := t.handler.ConnRead(conn, t.buf[:n])
		if data != nil {
			t.writeConn(conn, data)
		}
	}
}

func (t *eventLoop) handleConnWrite(conn *conn) {
	if _, err := conn.out.ReadToFd(conn.fd); err != nil {
		if err != unix.EAGAIN && err != unix.EINTR {
			t.closeConn(conn)
			return
		}
	}

	if conn.out.Len() == 0 {
		if err := t.poller.DisableWrite(conn.fd); err != nil {
			t.closeConn(conn)
			return
		}
	}
}

func (t *eventLoop) handleTask() {
	if t.tasks.IsEmpty() {
		return
	}

	log.Printf("%s process user tasks", t)

	for !t.tasks.IsEmpty() {
		if fn := t.tasks.Dequeue(); fn != nil {
			if err := fn(); err != nil {
				log.Printf("user action failed: %v", err)
			}
		}
	}

	if t.tasks.IsEmpty() {
		atomic.StoreInt32(&t.wakeup, 0)
	} else {
		if err := t.poller.Wakeup(); err != nil {
			if err == unix.EAGAIN || err == unix.EINTR {
				return
			}
			log.Fatalf("wakeup eventloop failed: %v", err)
		}
	}
}

func (t *eventLoop) closeConn(conn *conn) {
	t.poller.Delete(conn.fd)
}

func (t *eventLoop) writeConn(conn *conn, data []byte) error {
	var n int
	var err error
	if conn.out.Len() == 0 {
		if n, err = unix.Write(conn.fd, data); err != nil {
			if err != unix.EAGAIN && err != unix.EINTR {
				t.closeConn(conn)
				return err
			}
		}
	}

	if n > 0 {
		data = data[n:]
	}

	if n < len(data) {
		if _, err = conn.out.Write(data[n:]); err != nil {
			t.closeConn(conn)
			return err
		}
		if err = t.poller.EnableWrite(conn.fd); err != nil {
			t.closeConn(conn)
			return err
		}
	}

	return nil
}

func newEventLoop(handler EventHandler) (*eventLoop, error) {
	poller, err := poll.NewPoller()
	if err != nil {
		return nil, err
	}

	el := &eventLoop{
		id:      atomic.AddUint64(&eventLoopId, 1),
		poller:  poller,
		connMap: make(map[int]*conn),
		handler: handler,
		buf:     make([]byte, 4096),
		tasks:   newTaskQueue(),
	}

	return el, nil
}
