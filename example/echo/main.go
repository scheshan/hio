package main

import (
	"github.com/scheshan/hio"
	"log"
	"time"
)

func main() {
	handler := hio.EventHandler{
		SessionCreated: func(conn *hio.Conn) {
			log.Printf("connection[%v] created, bind to EventLoop[%v]", conn, conn.EventLoop().Id())
		},
		SessionRead: func(conn *hio.Conn, data []byte) []byte {
			log.Printf("connection[%v] receive data: %v", conn, len(data))
			return data
		},
		SessionClosed: func(conn *hio.Conn) {
			log.Printf("connection[%v] closed", conn)
		},
	}

	srv, err := hio.Serve("16379", handler)
	if err != nil {
		log.Fatal(err)
	}

	<-time.After(1 * time.Hour)

	srv.Shutdown()
}
