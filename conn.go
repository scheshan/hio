package hio

import (
	"errors"
	"fmt"
	"syscall"
)

type ConnState int

const (
	ConnState_Open ConnState = iota
	ConnState_Close
)

type Conn struct {
	id    int
	loop  *EventLoop
	fd    int
	addr  syscall.Sockaddr
	in    *Buffer
	out   *Buffer
	state ConnState
	attrs map[string]interface{}
}

func (t *Conn) Write(data []byte) error {
	if t.state == ConnState_Close {
		return errors.New("can't write to a closed connection")
	}

	t.out.Write(data)

	return nil
}

func (t *Conn) Close() {
	if t.state == ConnState_Close {
		return
	}

	t.state = ConnState_Close
	t.loop.tryCloseConn(t)
}

func (t *Conn) String() string {
	return fmt.Sprintf("%v", t.id)
}

func (t *Conn) Id() int {
	return t.id
}

func (t *Conn) State() ConnState {
	return t.state
}

func (t *Conn) Attr(key string) (interface{}, bool) {
	v, b := t.attrs[key]
	return v, b
}

func (t *Conn) SetAttr(key string, value interface{}) {
	if value == nil {
		delete(t.attrs, key)
	} else {
		t.attrs[key] = value
	}
}

func (t *Conn) EventLoop() *EventLoop {
	return t.loop
}

func newConn(loop *EventLoop, id int, fd int, addr syscall.Sockaddr) *Conn {
	bufIn := &Buffer{
		mp: loop.mp,
	}
	bufOut := &Buffer{
		mp: loop.mp,
	}

	conn := &Conn{
		id:    id,
		loop:  loop,
		fd:    fd,
		addr:  addr,
		in:    bufIn,
		out:   bufOut,
		attrs: make(map[string]interface{}),
	}

	return conn
}
