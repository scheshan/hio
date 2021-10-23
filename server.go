package hio

import (
	"errors"
	"log"
	"sync/atomic"
	"syscall"
)

type ConnCallback func(conn *Conn, data []byte, n int)

type ConnCallbackType int

type Server struct {
	port     int
	loops    []*EventLoop
	lb       LoadBalancer
	fd       int
	running  int32
	connId   int
	listener *Listener
}

func (t *Server) Run() error {
	if !atomic.CompareAndSwapInt32(&t.running, 0, 1) {
		return errors.New("server is already running")
	}

	if err := t.listen(); err != nil {
		return err
	}

	go t.accept()
	for _, loop := range t.loops {
		go loop.loop()
	}

	return nil
}

func (t *Server) listen() error {
	var fd int
	var err error

	if fd, err = syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0); err != nil {
		return err
	}

	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return err
	}
	if err = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1); err != nil {
		return err
	}

	sa := &syscall.SockaddrInet4{
		Port: t.port,
	}
	if err = syscall.Bind(fd, sa); err != nil {
		return err
	}

	if err = syscall.Listen(fd, 1024); err != nil {
		return err
	}

	t.fd = fd
	return nil
}

func (t *Server) isRunning() bool {
	return t.running == 1
}

func (t *Server) accept() {
	for t.isRunning() {
		fd, addr, err := syscall.Accept(t.fd)
		if err != nil {
			log.Fatal(err)
		}

		t.connId++

		loop := t.lb.Choose()
		loop.accept(t.connId, fd, addr)
	}
}

func (t *Server) initEventLoop(num int) {
	t.loops = make([]*EventLoop, num, num)
	for i := 0; i < num; i++ {
		t.loops[i] = newEventLoop(t, i)
	}
}

func NewServer() *Server {
	srv := &Server{}
	srv.port = 6379
	srv.initEventLoop(1)
	srv.lb = &lbRoundRobin{}
	srv.lb.Initialize(srv.loops)

	return srv
}
