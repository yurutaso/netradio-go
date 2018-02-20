package main

import (
	"fmt"
	"github.com/yurutaso/radiko-go"
	"log"
)

func main() {
	progs, err := radiko.GetStationProgramDate(`QRR`, 2018, 2, 15)
	if err != nil {
		log.Fatal(err)
	}

	progs, err = radiko.FilterByString(progs, `田村ゆかり`, `person`)
	if err != nil {
		log.Fatal(err)
	}
	if len(progs) == 0 {
		log.Fatal(fmt.Errorf(`No program found.`))
	}
	p := progs[0]

	radiko.Download("", p)
}
