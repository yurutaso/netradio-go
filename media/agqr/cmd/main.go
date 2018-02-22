package main

import (
	"github.com/yurutaso/netradio-go/media/agqr"
	"log"
)

func main() {
	err := agqr.Download(`test.m4a`, `1m`)
	if err != nil {
		log.Fatal(err)
	}
}
