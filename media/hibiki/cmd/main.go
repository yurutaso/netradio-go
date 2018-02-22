package main

import (
	"github.com/yurutaso/netradio-go/media/hibiki"
	"log"
)

func main() {
	/*
		stations, err := hibiki.GetStations()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(stations)
	*/
	prog, err := hibiki.GetProgram(`wug`)
	if err != nil {
		log.Fatal(err)
	}
	hibiki.Download(prog, ``)
}
