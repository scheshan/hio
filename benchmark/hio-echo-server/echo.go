package main

import (
	"flag"
	"github.com/scheshan/hio"
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

	handler := hio.EventHandler{
		SessionRead: func(conn *hio.Conn, data []byte) []byte {
			return data
		},
	}

	hio.ServeWithOptions(strconv.Itoa(port), handler, hio.ServerOptions{
		LoadBalancer: &hio.LoadBalancerRoundRobin{},
		EventLoopNum: loops,
	})

	<-time.After(1 * time.Hour)
}
