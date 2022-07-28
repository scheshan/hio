package hio

import (
	"github.com/scheshan/poll"
	"golang.org/x/sys/unix"
	"net"
	"sync/atomic"
)

type tcpServer struct {
	handler EventHandler
	options *Options
	domain  int
	sa      unix.Sockaddr
	poller  *poll.Poller
	lfd     int
	state   int32
}

func (t *tcpServer) Serve() error {
	if err := t.createPoller(); err != nil {
		return err
	}
	defer t.poller.Close()

	if err := t.createSocket(); err != nil {
		return err
	}
	defer unix.Close(t.lfd)

	if err := t.configureSocket(); err != nil {
		return err
	}

	defer t.closeEventLoops()
	if err := t.createEventLoops(); err != nil {
		return err
	}

	return t.accept()
}

func (t *tcpServer) Shutdown() {
	if !atomic.CompareAndSwapInt32(&t.state, 0, -1) {
		return
	}

	t.poller.Wakeup()
}

func (t *tcpServer) createPoller() error {
	poller, err := poll.NewPoller()
	if err != nil {
		return err
	}

	t.poller = poller
	return nil
}

func (t *tcpServer) createSocket() error {
	fd, err := unix.Socket(t.domain, unix.SOCK_STREAM, unix.IPPROTO_TCP)
	if err != nil {
		return err
	}

	t.lfd = fd
	return nil
}

func (t *tcpServer) configureSocket() error {
	opt := t.options

	if err := unix.SetNonblock(t.lfd, true); err != nil {
		return err
	}
	if opt.TcpNoDelay {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.TCP_NODELAY, 1); err != nil {
			return err
		}
	}
	if opt.ReuseAddr {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
			return err
		}
	}
	if opt.ReusePort {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
			return err
		}
	}
	if opt.RcvBuf > 0 {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_RCVBUF, int(opt.RcvBuf)); err != nil {
			return err
		}
	}
	if opt.SndBuf > 0 {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_SNDBUF, int(opt.SndBuf)); err != nil {
			return err
		}
	}
	if err := unix.Bind(t.lfd, t.sa); err != nil {
		return err
	}
	if err := unix.Listen(t.lfd, 1024); err != nil {
		return err
	}
	if err := t.poller.Add(t.lfd); err != nil {
		return err
	}

	return nil
}

func (t *tcpServer) closeEventLoops() {

}

func (t *tcpServer) createEventLoops() error {
	return nil
}

func (t *tcpServer) accept() error {
	for t.state == 0 {
		err := t.poller.Wait(30000, t.accept0)
		switch err {
		case nil, unix.EAGAIN, unix.EINTR:
			continue
		default:
			return err
		}
	}

	return nil
}

func (t *tcpServer) accept0(fd int, flag poll.Flag) error {
	cfd, _, err := unix.Accept(t.lfd)
	if err != nil {
		return err
	}

	if err := unix.SetNonblock(cfd, true); err != nil {
		unix.Close(cfd)
		return nil
	}
	if t.options.RcvBuf > 0 {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_RCVBUF, int(t.options.RcvBuf)); err != nil {
			return err
		}
	}
	if t.options.SndBuf > 0 {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.SO_SNDBUF, int(t.options.SndBuf)); err != nil {
			return err
		}
	}
	if t.options.TcpNoDelay {
		if err := unix.SetsockoptInt(t.lfd, unix.SOL_SOCKET, unix.TCP_NODELAY, 1); err != nil {
			return err
		}
	}

	unix.Write(cfd, []byte("hello world\r\n"))
	unix.Close(cfd)

	return nil
}

func newTcpServer(handler EventHandler, options *Options, addr *net.TCPAddr) *tcpServer {
	srv := &tcpServer{
		handler: handler,
		options: options,
	}
	srv.domain, srv.sa = resolveIpAndPort(addr.IP, addr.Port)

	return srv
}
