package hio

import (
	"github.com/scheshan/poll"
	"golang.org/x/sys/unix"
	"net"
	"strings"
)

type listener struct {
	network string
	domain  int
	typ     int
	proto   int
	sa      unix.Sockaddr
}

func (t *listener) setTcpAddr(proto string, addr *net.TCPAddr) {
	t.network = "tcp"
	t.typ = unix.SOCK_STREAM
	t.proto = unix.IPPROTO_TCP
	t.setIpAndPort(addr.IP, addr.Port)
}

func (t *listener) setUdpAddr(proto string, addr *net.UDPAddr) {
	t.network = "udp"
	t.typ = unix.SOCK_DGRAM
	t.proto = unix.IPPROTO_UDP
	t.setIpAndPort(addr.IP, addr.Port)
}

func (t *listener) setIpAndPort(ip net.IP, port int) {
	if ip.To4() != nil {
		t.domain = unix.AF_INET
		sa := &unix.SockaddrInet4{}
		sa.Port = port
		copy(sa.Addr[:], ip[12:16])
		t.sa = sa
	} else {
		t.domain = unix.AF_INET
		sa := &unix.SockaddrInet6{}
		sa.Port = port
		copy(sa.Addr[:], ip)
		t.sa = sa
	}
}

func (t *listener) setUnixAddr(proto string, addr *net.UnixAddr) {
	t.domain = unix.AF_LOCAL
	t.proto = 0
	sa := &unix.SockaddrUnix{}
	sa.Name = addr.Name
	t.sa = sa

	if proto == "unix" {
		t.network = "tcp"
		t.typ = unix.SOCK_STREAM
	} else {
		t.network = "udp"
		t.typ = unix.SOCK_DGRAM
	}
}

type server struct {
	handler  EventHandler
	options  *Options
	listener *listener
	lfd      int
	state    int32
	poller   *poll.Poller
}

func (t *server) Serve(addr string) error {
	l, err := t.createListener(addr)
	if err != nil {
		return err
	}
	t.listener = l

	poller, err := poll.NewPoller()
	if err != nil {
		return err
	}
	t.poller = poller
	defer poller.Close()

	defer func() {
		if t.lfd > 0 {
			unix.Close(t.lfd)
		}
	}()
	if err := t.configureSocket(); err != nil {
		return err
	}

	return t.accept()
}

func (t *server) createListener(addr string) (l *listener, err error) {
	l = &listener{}

	proto := "tcp"
	if i := strings.Index(addr, "://"); i > -1 {
		proto = addr[:i]
		addr = addr[i+3:]
	}

	switch proto {
	case "tcp", "tcp4", "tcp6":
		na, e := net.ResolveTCPAddr(proto, addr)
		if e != nil {
			return nil, e
		}
		l.setTcpAddr(proto, na)
	case "udp", "udp4", "udp6":
		na, e := net.ResolveUDPAddr(proto, addr)
		if e != nil {
			return nil, e
		}
		l.setUdpAddr(proto, na)
	case "unix", "unixgram":
		na, e := net.ResolveUnixAddr(proto, addr)
		if e != nil {
			return nil, e
		}
		l.setUnixAddr(proto, na)
	}

	return l, nil
}

func (t *server) configureSocket() error {
	fd, err := unix.Socket(t.listener.domain, t.listener.typ, t.listener.proto)
	if err != nil {
		return err
	}
	t.lfd = fd

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
	if err := unix.Bind(t.lfd, t.listener.sa); err != nil {
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

func (t *server) accept() error {
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

func (t *server) accept0(fd int, flag poll.Flag) error {
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
