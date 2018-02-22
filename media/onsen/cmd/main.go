package main

import (
	"github.com/yurutaso/netradio-go/media/onsen"
	"log"
)

func main() {
	prog, err := onsen.GetProgram(`wug`)
	if err != nil {
		log.Fatal(err)
	}
	onsen.Download(prog, ``)
}
