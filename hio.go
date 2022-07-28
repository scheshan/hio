package hio

import (
	"errors"
	"golang.org/x/sys/unix"
	"net"
	"runtime"
	"strings"
)

type EventHandler struct {
	Boot       func()
	ConnCreate func(conn Conn)
	ConnRead   func(conn Conn, data []byte) []byte
	ConnClose  func(conn Conn)
}

type Options struct {
	EventLoopNum int
	ReuseAddr    bool
	ReusePort    bool
	TcpNoDelay   bool
	SndBuf       uint64
	RcvBuf       uint64
}

type OptionsFunc func(opt *Options)

func Serve(addr string, handler EventHandler, optFunc ...OptionsFunc) error {
	options := &Options{
		EventLoopNum: runtime.NumCPU(),
		RcvBuf:       4096,
		SndBuf:       4096,
	}
	for _, f := range optFunc {
		f(options)
	}

	srv, err := createServer(handler, options, addr)
	if err != nil {
		return err
	}

	return srv.Serve()
}

func createServer(handler EventHandler, options *Options, addr string) (Server, error) {
	proto := "tcp"
	if i := strings.Index(addr, "://"); i > -1 {
		proto = addr[:i]
		addr = addr[i+3:]
	}

	switch proto {
	case "tcp", "tcp4", "tcp6":
		na, err := net.ResolveTCPAddr(proto, addr)
		if err != nil {
			return nil, err
		}
		return newTcpServer(handler, options, na), nil
	case "udp", "udp4", "udp6":
		fallthrough
		//na, err := net.ResolveUDPAddr(proto, addr)
		//if err != nil {
		//	return nil, err
		//}
		//na = nil
	case "unix":
		fallthrough
		//na, err := net.ResolveUnixAddr(proto, addr)
		//if err != nil {
		//	return nil, err
		//}
	default:
		return nil, errors.New("invalid addr")
	}
}

func resolveIpAndPort(ip net.IP, port int) (domain int, sa unix.Sockaddr) {
	if ip.To4() == nil {
		raw := &unix.SockaddrInet6{}
		raw.Port = port
		copy(raw.Addr[:], ip)

		return unix.AF_INET6, raw
	} else {
		raw := &unix.SockaddrInet4{}
		raw.Port = port
		copy(raw.Addr[:], ip[12:16])

		return unix.AF_INET, raw
	}
}
