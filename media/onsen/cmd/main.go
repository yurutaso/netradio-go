package main

import (
	"fmt"
	"github.com/yurutaso/go-anirad/media/onsen"
	"log"
)

func main() {
	prog, err := onsen.GetProgram(`wug`)
	if err != nil {
		log.Fatal(err)
	}
	onsen.Download(prog, ``)
}
