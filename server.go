package hio

import (
	"errors"
	"net"
	"strings"
	"syscall"
)

type Server interface {
	Shutdown() error
}

type ServerOptions struct {
	EventLoopNum int
}

func resolveIpAndPort(ip net.IP, port int) syscall.Sockaddr {
	var sa syscall.Sockaddr
	var addr []byte

	if len(ip) == net.IPv4len {
		sa4 := &syscall.SockaddrInet4{}
		sa4.Port = port
		addr = sa4.Addr[:]
		sa = sa4
	} else {
		sa6 := &syscall.SockaddrInet6{}
		sa6.Port = port
		addr = sa6.Addr[:]
		sa = sa6
	}

	for i := 0; i < len(addr); i++ {
		addr[i] = ip[i]
	}

	return sa
}

func resolveAddr(addr string) (proto string, sa syscall.Sockaddr, err error) {
	proto = "tcp"
	if strings.Contains(addr, "://") {
		proto = strings.Split(addr, "://")[0]
	}

	sa = nil

	switch proto {
	case "tcp":
		tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
		if err != nil {
			return proto, nil, err
		}
		sa = resolveIpAndPort(tcpAddr.IP, tcpAddr.Port)
		return
	case "udp":
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return
		}
		sa = resolveIpAndPort(udpAddr.IP, udpAddr.Port)
		return
	case "unix":
		unixAddr, err := net.ResolveUnixAddr("unix", addr)
		if err != nil {
			return
		}

		sa := &syscall.SockaddrUnix{}
		sa.Name = unixAddr.Name
		return
	default:
		err = errors.New("Protocol " + proto + " not supported")
		return
	}
}

func Serve(addr string, opt *ServerOptions) (Server, error) {
	if opt == nil {
		return nil, errors.New("opt is required")
	}

	proto, sa, err := resolveAddr(addr)
	if err != nil {
		return nil, err
	}

	switch proto {
	case "tcp":
		return serveTcp(sa, opt)
	default:
		return nil, errors.New("Protocol " + proto + " not supported")
	}
}

func serveTcp(sa syscall.Sockaddr, opt *ServerOptions) (Server, error) {
	srv := &tcpServer{}
	srv.addr = sa

	if err := srv.run(); err != nil {
		return nil, err
	}
	return srv, nil
}
