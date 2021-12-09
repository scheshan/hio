package main

import (
	"github.com/scheshan/hio"
	"github.com/scheshan/hio/buf"
	"log"
	"time"
)

func main() {
	opt := hio.ServerOptions{}
	opt.EventLoopNum = 4
	opt.OnSessionCreated = func(conn *hio.Conn) {
		log.Printf("connection[%v] created, bind to EventLoop[%v]", conn, conn.EventLoop().Id())
	}
	opt.OnSessionRead = func(conn *hio.Conn, buffer *buf.Buffer) *buf.Buffer {
		log.Printf("connection[%v] receive data: %v", conn, buffer.ReadableBytes())
		buffer.IncrRef()
		return buffer
	}
	opt.OnSessionClosed = func(conn *hio.Conn) {
		log.Printf("connection[%v] closed", conn)
	}

	srv, err := hio.Serve("16379", opt)
	if err != nil {
		log.Fatal(err)
	}

	<-time.After(1 * time.Hour)

	srv.Shutdown()
}
