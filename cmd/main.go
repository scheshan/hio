package main

import (
	"hio"
	"log"
	"time"
)

func main() {
	srv := hio.NewServer()
	err := srv.Run()

	if err != nil {
		log.Fatal(err)
	}

	<-time.After(1000 * time.Hour)
}
