package hio

import (
	"errors"
	"log"
	"sync/atomic"
	"syscall"
)

type tcpServer struct {
	addr    syscall.Sockaddr
	nw      *network
	running int32
	lfd     int
}

func (t *tcpServer) run() error {
	if !atomic.CompareAndSwapInt32(&t.running, 0, 1) {
		return errors.New("server already running")
	}

	if err := t.bindAndListen(); err != nil {
		return err
	}

	if err := t.initNetwork(); err != nil {
		return err
	}

	go t.loop()

	return nil
}

func (t *tcpServer) bindAndListen() error {
	var soType int
	if _, ok := t.addr.(*syscall.SockaddrInet4); ok {
		soType = syscall.AF_INET
	} else {
		soType = syscall.AF_INET6
	}

	fd, err := syscall.Socket(soType, syscall.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	t.lfd = fd

	if err = syscall.SetsockoptInt(t.lfd, 0, syscall.SO_REUSEADDR, 1); err != nil {
		return err
	}

	if err = syscall.Bind(t.lfd, t.addr); err != nil {
		return err
	}

	if err = syscall.Listen(t.lfd, 1024); err != nil {
		return err
	}

	log.Printf("server started and listen on addr: %v", t.addr)

	return nil
}

func (t *tcpServer) initNetwork() error {
	nw, err := newNetwork()
	if err != nil {
		return err
	}
	t.nw = nw

	if err := t.nw.addRead(t.lfd); err != nil {
		return err
	}

	return nil
}

func (t *tcpServer) loop() {
	for t.running == 1 {
		n, err := t.nw.wait(networkWaitMs)
		if err != nil {
			if err == syscall.EAGAIN {
				continue
			}
			t.shutdown()
			return
		}

		for range n {
			fd, _, err := syscall.Accept(t.lfd)
			if err != nil {
				if err == syscall.EAGAIN {
					continue
				}
				t.shutdown()
				return
			}

			log.Printf("new connection: %v, now close it", fd)
			syscall.Close(fd)
		}

		log.Print(n)
	}
}

func (t *tcpServer) shutdown() {
	t.running = 0
	if t.nw != nil {
		t.nw.shutdown()
		t.nw = nil
	}
	if t.lfd > 0 {
		syscall.Close(t.lfd)
		t.lfd = 0
	}
}

func (t *tcpServer) Shutdown() error {
	if !atomic.CompareAndSwapInt32(&t.running, 1, 0) {
		return errors.New("server not running")
	}

	t.shutdown()
	return nil
}
