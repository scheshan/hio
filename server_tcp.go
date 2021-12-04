package hio

import (
	"errors"
	"log"
	"runtime"
	"sync/atomic"
	"syscall"
)

type tcpServer struct {
	addr    syscall.Sockaddr
	nw      *network
	running int32
	lfd     int
	connId  uint64
	loops   []*EventLoop
	opt     ServerOptions
}

func (t *tcpServer) run() error {
	if !atomic.CompareAndSwapInt32(&t.running, 0, 1) {
		return errors.New("server already running")
	}

	if err := t.bindAndListen(); err != nil {
		return err
	}

	if err := t.initEventLoops(); err != nil {
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

	if err = syscall.SetsockoptInt(t.lfd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1); err != nil {
		return err
	}

	if err = syscall.SetsockoptInt(t.lfd, syscall.SOL_SOCKET, syscall.SO_REUSEPORT, 1); err != nil {
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

func (t *tcpServer) initEventLoops() error {
	loopNum := t.opt.EventLoopNum
	if loopNum <= 0 {
		loopNum = runtime.NumCPU()
	}

	t.loops = make([]*EventLoop, loopNum)
	for i := 0; i < loopNum; i++ {
		el, err := newEventLoop(uint64(i), t.opt)
		if err != nil {
			return err
		}

		t.loops[i] = el
		el.run()
	}

	return nil
}

func (t *tcpServer) loop() {
	for t.running == 1 {
		n, err := t.nw.wait(networkWaitMs)
		if err != nil {
			if err == syscall.EAGAIN || err == syscall.EINTR {
				continue
			}
			t.shutdown()
			return
		}

		for range n {
			fd, sa, err := syscall.Accept(t.lfd)
			if err != nil {
				if err == syscall.EAGAIN || err == syscall.EINTR {
					continue
				}
				t.shutdown()
				return
			}

			t.connId++

			conn := newConn(t.connId, sa, fd)
			t.handleNewConn(conn)
		}

		log.Print(n)
	}
}

func (t *tcpServer) handleNewConn(conn *Conn) {
	//TODO load balance
	el := t.loops[0]
	el.addConn(conn)
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
	if t.loops != nil {
		loops := t.loops
		t.loops = nil

		for _, loop := range loops {
			if loop != nil {
				loop.shutdown()
			}
		}
	}
}

func (t *tcpServer) Shutdown() error {
	if !atomic.CompareAndSwapInt32(&t.running, 1, 0) {
		return errors.New("server not running")
	}

	t.shutdown()
	return nil
}
