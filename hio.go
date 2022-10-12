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
	EventLoopNum   int
	ReuseAddr      bool
	ReusePort      bool
	TcpNoDelay     bool
	TcpSndBuf      uint64
	TcpRcvBuf      uint64
	WriteBufferCap uint64
	ReadBufferSize uint64
	LB             LoadBalancerType
}

func (t *Options) validate() error {
	return nil
}

type OptionsFunc func(opt *Options)

func Serve(addr string, handler EventHandler, optFunc ...OptionsFunc) error {
	options := &Options{
		EventLoopNum:   runtime.NumCPU(),
		TcpRcvBuf:      4096,
		TcpSndBuf:      4096,
		LB:             RoundRobin,
		ReadBufferSize: 4096,
		WriteBufferCap: 1024 * 1024,
	}
	for _, f := range optFunc {
		f(options)
	}

	if err := options.validate(); err != nil {
		return err
	}

	srv := new(server)
	srv.options = options
	srv.handler = handler

	return srv.Serve(addr)
}
