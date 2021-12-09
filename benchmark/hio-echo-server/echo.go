package main

import (
	"flag"
	"github.com/scheshan/hio"
	"github.com/scheshan/hio/buf"
	"strconv"
	"time"
)

func main() {
	//go func() {
	//	if err := http.ListenAndServe(":6089", nil); err != nil {
	//		panic(err)
	//	}
	//}()

	var port int
	var loops int

	flag.IntVar(&port, "port", 1833, "server port")
	flag.IntVar(&loops, "loops", -1, "num loops")
	flag.Parse()

	hio.Serve(strconv.Itoa(port), hio.ServerOptions{
		LoadBalancer:     &hio.LoadBalancerRoundRobin{},
		EventLoopNum:     loops,
		OnSessionCreated: nil,
		OnSessionRead: func(conn *hio.Conn, buffer *buf.Buffer) {
			conn.Write(buffer)
		},
		OnSessionClosed: nil,
	})

	<-time.After(1 * time.Hour)
}
