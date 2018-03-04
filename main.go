package main

import (
	"flag"
	"fmt"
	"github.com/yurutaso/netradio-go/media/agqr"
	"github.com/yurutaso/netradio-go/media/hibiki"
	"github.com/yurutaso/netradio-go/media/onsen"
	"github.com/yurutaso/netradio-go/media/radiko"
	"log"
	"os"
	"path/filepath"
	"time"
)
const (
	HELP string = `
	Usage: onsen [-l] -s station [-i] [-o output]
	Usage: hibiki [-l] -s station [-i] [-o output]
	Usage: radiko [-l] -s station [selections "-n name" "-p person" "-i info" of the program]
	Usage: agqr -o output -d duration(default: 30m)
	`
)

func downloadRadiko(title, person, info, station string) error {
	progs, err := radiko.GetStationProgramWeek(station)
	if err != nil {
		return err
	}
	progs, err = radiko.FilterByString(progs, title, `title`)
	if err != nil {
		return err
	}
	progs, err = radiko.FilterByString(progs, person, `person`)
	if err != nil {
		return err
	}
	progs, err = radiko.FilterByString(progs, info, `info`)
	if err != nil {
		return err
	}
	if len(progs) == 0 {
		return fmt.Errorf(`No radio program found`)
	}
	prog := progs[0]
	radiko.Download(``, prog)
	return nil
}

func downloadOnsen(station, fileout string) error {
	prog, err := onsen.GetProgram(station)
	if err != nil {
		return err
	}
	return onsen.Download(prog, fileout)
}

func downloadHibiki(station, fileout string) error {
	prog, err := hibiki.GetProgram(station)
	if err != nil {
		return err
	}
	return hibiki.Download(prog, fileout)
}

func downloadAGQR(fileout, duration string) error {
	return agqr.Download(fileout, duration)
}

func listOnsen() error {
	stations, err := onsen.GetStations()
	if err != nil {
		return err
	}
	fmt.Println(stations)
	return nil
}

func listHibiki() error {
	stations, err := hibiki.GetStations()
	if err != nil {
		return err
	}
	fmt.Println(stations)
	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, HELP)
		flag.PrintDefaults()
	}
	flag.Parse()

	if len(os.Args) == 1 {
		fmt.Printf(HELP)
		return
	}
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, HELP)
		flag.PrintDefaults()
	}
	var (
		optD   = fs.String("d", "", "description of the program to filter")
		optDIR = fs.String("dir", "", "output directory (ignored if -o is set)")
		optO   = fs.String("o", "", "output file")
		optN   = fs.String("n", "", "name of the title to filter")
		optP   = fs.String("p", "", "person to filter")
		optS   = fs.String("s", "", "station name")
		optT   = fs.String("t", "30m", "time duration to record AGQR(default:30m)")
		flagI  = fs.Bool("i", false, "show info of a program. (ignored if -l is set)")
		flagL  = fs.Bool("l", false, "list stations.")
	)
	fs.Parse(os.Args[2:])

	var err error = nil
	switch os.Args[1] {
	case `onsen`:
		if *flagL {
			err = listOnsen()
			break
		}
		station := *optS
		if station == `` {
			log.Fatal(fmt.Errorf(`Error! You must set station with -s.`))
		}
		if *flagI {
			prog, err := onsen.GetProgram(station)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(prog)
			break
		}
		err = downloadOnsen(station, *optO)
	case `hibiki`:
		if *flagL {
			err = listHibiki()
			break
		}
		station := *optS
		if station == `` {
			log.Fatal(fmt.Errorf(`Error! You must set station with -s.`))
		}
		if *flagI {
			prog, err := hibiki.GetProgram(station)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(prog)
			break
		}
		err = downloadHibiki(station, *optO)
	case `radiko`:
		if *flagL {
			err = radiko.ListStations(radiko.AREA_TABLE[`Tokyo`])
			if err != nil {
				log.Fatal(err)
			}
			break
		}
		if *flagI {
			log.Fatal(fmt.Errorf(`Error! Invalid option -i with radiko.`))
		}
		if *optO != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -o with radiko.`))
		}
		if *optDIR != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -dir with radiko.`))
		}
		if *optT {
			log.Fatal(fmt.Errorf(`Error! Invalid option -t with radiko.`))
		}
		station := *optS
		if station == `` {
			log.Fatal(fmt.Errorf(`Error! You must set station with -s.`))
		}
		title := *optN
		person := *optP
		description := *optD
		err = downloadRadiko(title, person, description, station)
	case `agqr`:
		if *flagI {
			log.Fatal(fmt.Errorf(`Error! Invalid option -i with agqr.`))
		}
		if *flagL {
			log.Fatal(fmt.Errorf(`Error! Invalid option -l with agqr.`))
		}
		if *optD != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -d with agqr.`))
		}
		if *optO != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -o with agqr.`))
		}
		duration := *optT
		t := time.Now()
		fileout := filepath.Join(*optDIR, fmt.Sprintf("%4d%02d%02d%02d%02d_AGQR.flv", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute()))
		err = downloadAGQR(fileout, duration)
	default:
		log.Fatal(fmt.Errorf(`Invalid media (onsen/hibiki/radiko/agqr)`))
	}
	if err != nil {
		log.Fatal(err)
	}
}
