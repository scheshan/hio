package hio

import (
	"fmt"
	"golang.org/x/sys/unix"
)

type Conn interface {
	fmt.Stringer
	Id() uint64
}

type tcpConn struct {
	id uint64
	sa unix.Sockaddr
	fd int
}

func (t *tcpConn) String() string {
	return ""
}

func (t *tcpConn) Id() uint64 {
	return t.id
}
