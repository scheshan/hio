package main

import (
	"hio"
	"log"
	"time"
)

func main() {
	opts := &hio.ServerOptions{}
	opts.Port = 6379
	opts.Listener = &hio.Listener{
		OnConnRead: func(conn *hio.Conn, data []byte, n int) {
			conn.Write(data[0:n])
		},
	}

	srv := hio.NewServer(opts)
	err := srv.Run()

	if err != nil {
		log.Fatal(err)
	}

	<-time.After(1000 * time.Hour)
}
