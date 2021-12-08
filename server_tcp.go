package hio

import (
	"errors"
	"github.com/scheshan/hio/poll"
	"golang.org/x/sys/unix"
	"log"
	"runtime"
	"sync/atomic"
)

type tcpServer struct {
	addr    unix.Sockaddr
	poller  *poll.Poller
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

	if err := t.initPoller(); err != nil {
		return err
	}

	go t.loop()

	return nil
}

func (t *tcpServer) bindAndListen() error {
	var soType int
	if _, ok := t.addr.(*unix.SockaddrInet4); ok {
		soType = unix.AF_INET
	} else {
		soType = unix.AF_INET6
	}

	fd, err := unix.Socket(soType, unix.SOCK_STREAM, 0)
	if err != nil {
		return err
	}
	t.lfd = fd

	if err = unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		return err
	}

	if err = unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
		return err
	}

	if err = unix.Bind(t.lfd, t.addr); err != nil {
		return err
	}

	if err = unix.Listen(t.lfd, 1024); err != nil {
		return err
	}

	log.Printf("server started and listen on addr: %v", t.addr)

	return nil
}

func (t *tcpServer) initPoller() error {
	p, err := poll.NewPoller()
	if err != nil {
		return err
	}
	t.poller = p

	if err := t.poller.Add(t.lfd); err != nil {
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
		go el.loop()
	}

	return nil
}

func (t *tcpServer) loop() {
	for t.running == 1 {
		n, err := t.poller.Wait(poll.DefaultWaitMs)
		if err != nil {
			if err == unix.EAGAIN || err == unix.EINTR {
				continue
			}
			t.shutdown()
			return
		}

		for range n {
			fd, sa, err := unix.Accept(t.lfd)
			if err != nil {
				if err == unix.EAGAIN || err == unix.EINTR {
					continue
				}
				t.shutdown()
				return
			}

			t.connId++

			conn := newConn(t.connId, fd, sa)
			t.handleNewConn(conn)
		}
	}
}

func (t *tcpServer) handleNewConn(conn *Conn) {
	err := unix.SetNonblock(conn.fd, true)
	if err != nil {
		conn.release()
		return
	}

	//TODO load balance
	el := t.loops[0]
	el.bindConn(conn)
}

func (t *tcpServer) shutdown() {
	t.running = 0
	if t.poller != nil {
		t.poller.Close()
		t.poller = nil
	}
	if t.lfd > 0 {
		unix.Close(t.lfd)
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
