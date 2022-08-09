package hio

import (
	"runtime"
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
	LB           LoadBalancerType
}

type OptionsFunc func(opt *Options)

func Serve(addr string, handler EventHandler, optFunc ...OptionsFunc) error {
	options := &Options{
		EventLoopNum: runtime.NumCPU(),
		RcvBuf:       4096,
		SndBuf:       4096,
		LB:           RoundRobin,
	}
	for _, f := range optFunc {
		f(options)
	}

	srv := new(server)
	srv.options = options
	srv.handler = handler

	return srv.Serve(addr)
}
