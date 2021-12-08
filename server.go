package hio

import (
	"errors"
	"github.com/scheshan/hio/buf"
	"golang.org/x/sys/unix"
	"net"
	"strconv"
)

type Server interface {
	Shutdown() error
}

type ServerOptions struct {
	EventLoopNum     int
	OnSessionCreated func(conn *Conn)
	OnSessionRead    func(conn *Conn, buf *buf.Buffer)
	OnSessionClosed  func(conn *Conn)
}

func resolveIpAndPort(ip net.IP, port int) unix.Sockaddr {
	var sa unix.Sockaddr
	var addr []byte

	if ip == nil || len(ip) == net.IPv4len {
		sa4 := &unix.SockaddrInet4{}
		sa4.Port = port
		addr = sa4.Addr[:]
		sa = sa4
	} else {
		sa6 := &unix.SockaddrInet6{}
		sa6.Port = port
		addr = sa6.Addr[:]
		sa = sa6
	}

	for i := 0; i < len(ip); i++ {
		addr[i] = ip[i]
	}

	return sa
}

func resolveAddr(addr string) (proto string, sa unix.Sockaddr, err error) {
	port, err := strconv.Atoi(addr)
	if err != nil {
		return "", nil, err
	}

	proto = "tcp"
	sa = &unix.SockaddrInet4{
		Port: port,
		Addr: [4]byte{},
	}
	err = nil
	return

	//TODO: resolve real address
	//proto = "tcp"
	//if strings.Contains(addr, "://") {
	//	arr := strings.Split(addr, "://")
	//	proto = arr[0]
	//	addr = arr[1]
	//}
	//
	//sa = nil
	//
	//switch proto {
	//case "tcp":
	//	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	//	if err != nil {
	//		return proto, nil, err
	//	}
	//	sa = resolveIpAndPort(tcpAddr.IP, tcpAddr.Port)
	//	return proto, sa, nil
	//case "udp":
	//	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	//	if err != nil {
	//		return proto, nil, err
	//	}
	//	sa = resolveIpAndPort(udpAddr.IP, udpAddr.Port)
	//	return proto, sa, nil
	//case "unix":
	//	unixAddr, err := net.ResolveUnixAddr("unix", addr)
	//	if err != nil {
	//		return proto, nil, err
	//	}
	//
	//	sa := &syscall.SockaddrUnix{}
	//	sa.Name = unixAddr.Name
	//	return proto, sa, nil
	//default:
	//	err = errors.New("Protocol " + proto + " not supported")
	//	return proto, nil, err
	//}
}

func Serve(addr string, opt ServerOptions) (Server, error) {
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

func serveTcp(sa unix.Sockaddr, opt ServerOptions) (Server, error) {
	srv := &tcpServer{}
	srv.addr = sa
	srv.opt = opt

	if err := srv.run(); err != nil {
		return nil, err
	}
	return srv, nil
}
