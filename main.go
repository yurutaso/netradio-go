package main

import (
	"flag"
	"fmt"
	"github.com/yurutaso/netradio-go/media/agqr"
	"github.com/yurutaso/netradio-go/media/ann"
	"github.com/yurutaso/netradio-go/media/hibiki"
	"github.com/yurutaso/netradio-go/media/onsen"
	"github.com/yurutaso/netradio-go/media/radiko"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	HELP string = `
	Usage: onsen [-l] -s station [-i] [-o output]
	Usage: hibiki [-l] -s station [-i] [-o output]
	Usage: radiko [-l] -s station [selections "-n name" "-p person" "-d description" "-day day(e.g. 20180507)" of the program]
	Usage: agqr -o output -t time(default: 30m)
	Usage: ann [-u username] [-p password] [-all]
	`
)

func downloadRadiko(title, person, info, day, station string) error {
	var progs []radiko.Program
	var err error
	if len(day) == 0 {
		progs, err = radiko.GetStationProgramWeek(station)
		if err != nil {
			return err
		}
	} else {
		if len(day) != 8 {
			fmt.Println(day)
			return fmt.Errorf("opt -day must be numbers YYYYMMDD (e.g. 20180527)")
		}
		y, err := strconv.Atoi(day[0:4])
		if err != nil {
			return err
		}
		m, err := strconv.Atoi(day[4:6])
		if err != nil {
			return err
		}
		d, err := strconv.Atoi(day[6:8])
		if err != nil {
			return err
		}
		progs, err = radiko.GetStationProgramDate(station, y, m, d)
		if err != nil {
			return err
		}
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

func downloadANN(fileout string, id string, password string, getall bool) error {
	client := &http.Client{}
	if id != "" && password != "" {
		if err := ann.Login(client, id, password); err != nil {
			return err
		}
	}

	var progs []*ann.Program
	var err error
	if !getall {
		progs, err = ann.GetLatestProgram(client)
	} else {
		progs, err = ann.GetPrograms(client)
	}
	if err != nil {
		return err
	}

	for _, prog := range progs {
		if err := ann.Download(client, prog, ""); err != nil {
			return err
		}
	}
	return nil
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
		optD    = fs.String("d", "", "description of the program to filter")
		optDay  = fs.String("day", "", "day (e.g. 20180507)")
		optDIR  = fs.String("dir", "", "output directory (ignored if -o is set)")
		optO    = fs.String("o", "", "output file")
		optN    = fs.String("n", "", "name of the title to filter")
		optP    = fs.String("p", "", "password (ann), person to filter (radiko)")
		optS    = fs.String("s", "", "station name")
		optT    = fs.String("t", "", "time duration to record AGQR(default:30m)")
		optU    = fs.String("u", "", "username")
		flagI   = fs.Bool("i", false, "show info of a program. (ignored if -l is set)")
		flagL   = fs.Bool("l", false, "list stations.")
		flagAll = fs.Bool("all", false, "download all programs (ann only).")
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
		if *optT != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -t with radiko.`))
		}
		station := *optS
		if station == `` {
			log.Fatal(fmt.Errorf(`Error! You must set station with -s.`))
		}
		title := *optN
		person := *optP
		description := *optD
		day := *optDay
		err = downloadRadiko(title, person, description, day, station)
	case `ann`:
		if *flagI {
			log.Fatal(fmt.Errorf(`Error! Invalid option -i with ann.`))
		}
		if *optDIR != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -dir with ann.`))
		}
		if *flagL {
			log.Fatal(fmt.Errorf(`Error! Invalid option -l with ann.`))
		}
		if *optD != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -d with ann.`))
		}
		if *optT != "" {
			log.Fatal(fmt.Errorf(`Error! Invalid option -t with ann.`))
		}
		err = downloadANN(*optO, *optU, *optP, *flagAll)
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
		duration := ""
		if *optT == "" {
			duration = "30m"
		} else {
			duration = *optT
		}
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
