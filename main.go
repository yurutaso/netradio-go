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
	"time"
)

const (
	HELP string = `
	Usage: onsen [-l] -s station [-i] [-o output]
	Usage: hibiki [-l] -s station [-i] [-o output]
	Usage: radiko [-l] -s station [selections "-n name" "-p person" "-d description" "-day day(e.g. 20180507)" of the program]
	Usage: agqr -o output -t time(default: 30m)
	Usage: ann [-l] [-u username] [-p password] [-m maxnumber]
	`
)

func downloadRadiko(title, person, info, day, station string) error {
	progs, err := radiko.Find(title, person, info, day, station)
	if err != nil {
		return err
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

func downloadANN(client *http.Client, fileout string, max int) error {
	var progs []*ann.Program
	var err error
	if max < 1 {
		progs, err = ann.GetLatestProgram(client)
	} else {
		progs, err = ann.GetPrograms(client, max)
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
		optM    = fs.Int("m", -1, "max number (default -1: all programs)")
		optN    = fs.String("n", "", "name of the title to filter")
		optP    = fs.String("p", "", "password (ann), person to filter (radiko)")
		optS    = fs.String("s", "", "station name")
		optT    = fs.String("t", "", "time duration to record AGQR(default:30m)")
		optU    = fs.String("u", "", "username")
		flagI   = fs.Bool("i", false, "show info of a program. (ignored if -l is set)")
		flagL   = fs.Bool("l", false, "list stations.")
		flagDry = fs.Bool("dry", false, "dry-run the download. Only show the output name.")
	)
	fs.Parse(os.Args[2:])

	var err error = nil
	switch os.Args[1] {
	case `onsen`:
		if *flagL {
			err = listOnsen()
			break
		}
		if *optS == `` {
			log.Fatal(fmt.Errorf(`Error! You must set station with -s.`))
		}
		station := *optS
		if *flagI || *flagDry {
			prog, err := onsen.GetProgram(station)
			if err != nil {
				log.Fatal(err)
			}
			if *flagDry {
				s, err := onsen.GetOutputFilename(prog, *optO)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(s)
				break
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
		if *optS == `` {
			log.Fatal(fmt.Errorf(`Error! You must set station with -s.`))
		}
		station := *optS

		if *flagI || *flagDry {
			prog, err := hibiki.GetProgram(station)
			if err != nil {
				log.Fatal(err)
			}
			if *flagDry {
				s, err := hibiki.GetOutputFilename(prog, *optO)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(s)
				break
			}
			fmt.Println(prog)
			break
		}
		err = downloadHibiki(station, *optO)
	case `radiko`:
		if *optS == `` {
			log.Fatal(fmt.Errorf(`Error! You must set station with -s.`))
		}
		station := *optS
		if *flagL {
			err = radiko.ListStations(radiko.AREA_TABLE[`Tokyo`])
			if err != nil {
				log.Fatal(err)
			}
			break
		}
		title := *optN
		person := *optP
		description := *optD
		day := *optDay
		if *flagI || *flagDry {
			progs, err := radiko.Find(title, person, description, day, station)
			if err != nil {
				log.Fatal(err)
			}
			prog := progs[0]

			if *flagDry {
				s, err := radiko.GetOutputFilename(prog, *optO)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(s)
				break
			}
			fmt.Println(prog)
			break
		}
		err = downloadRadiko(title, person, description, day, station)
	case `ann`:
		client := &http.Client{}
		if *optU != "" && *optP != "" {
			if err := ann.Login(client, *optU, *optP); err != nil {
				log.Fatal(err)
			}
		}
		if *flagDry {
			progs, err := ann.GetLatestProgram(client)
			if err != nil {
				log.Fatal(err)
			}
			s, err := ann.GetOutputFilename(progs[0], *optO)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(s)
			break
		}
		err = downloadANN(client, *optO, *optM)
	case `agqr`:
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
		log.Fatal(fmt.Errorf(`Invalid media (onsen/hibiki/radiko/agqr/ann)`))
	}
	if err != nil {
		log.Fatal(err)
	}
}
