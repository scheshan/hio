package hio

import (
	"syscall"
)

type tcpServer struct {
	addr syscall.Sockaddr
}

func (t *tcpServer) run() error {
	return nil
}

func (t *tcpServer) Shutdown() error {
	return nil
}
