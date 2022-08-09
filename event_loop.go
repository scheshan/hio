package hio

import (
	"github.com/scheshan/poll"
	"golang.org/x/sys/unix"
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

func (t *eventLoop) callback(fd int, flag poll.Flag) error {
	conn, ok := t.connMap[fd]
	if !ok {
		return nil
	}

	if flag.CanRead() {
		t.readConn(conn)
	}
	if flag.CanWrite() {
		t.writeConn(conn)
	}

	return nil
}

func (t *eventLoop) readConn(conn *conn) {

}

func (t *eventLoop) writeConn(conn *conn) {

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
	}

	return el, nil
}
