package hio

import (
	"errors"
	"fmt"
	"github.com/scheshan/buffer"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"sync/atomic"
)

var connId uint64

type Conn interface {
	fmt.Stringer
	Id() uint64
	Close()
	Write(data []byte) error
	Active() bool
}

type conn struct {
	id    uint64
	sa    unix.Sockaddr
	fd    int
	loop  *eventLoop
	in    *buffer.Buffer
	out   *buffer.Buffer
	log   *log.Entry
	state int32 //0:active, -1:inactive, -2:closed
}

func (t *conn) String() string {
	return fmt.Sprintf("conn-%v", t.id)
}

func (t *conn) Id() uint64 {
	return t.id
}

func (t *conn) Close() {
	if !atomic.CompareAndSwapInt32(&t.state, 0, -1) {
		return
	}

	t.loop.AddTask(func() error {
		t.loop.closeConn(t, nil)
		return nil
	})
}

func (t *conn) Write(data []byte) error {
	if !t.Active() {
		return errors.New("conn is disconnected")
	}

	t.loop.AddTask(func() error {
		t.loop.writeConn(t, data)
		return nil
	})
	return nil
}

func (t *conn) Active() bool {
	return atomic.LoadInt32(&t.state) == 0
}

func newConn(fd int, sa unix.Sockaddr) *conn {
	c := &conn{
		id: atomic.AddUint64(&connId, 1),
		sa: sa,
		fd: fd,
	}
	return c
}
