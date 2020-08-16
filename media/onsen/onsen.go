package onsen

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	//"sort"
	"strings"
)

const (
	//ONSEN_PROGRAM_JSON = `http://www.onsen.ag/api/shownMovie/shownMovie.json`
	//ONSEN_PROGRAM_API  = `http://www.onsen.ag/data/api`
    ONSEN_PROGRAM_JSON = `https://www.onsen.ag/web_api/programs`
)

type Program struct {
	Station string
	Url     string
	Title   string
	Date    string
    Id int
}

// 番組内の各配信
type OnsenContent struct {
    Url string `json:"streaming_url"`
    Title string `json:"title"`
    Premium bool `json:"premium"`
    Date string `json:"delivery_date"`
    Latest bool `json:"latest"`
    Id int `json:"id"`
}

type OnsenProgramInfo struct {
    Title string
}

// 一つの番組
type OnsenProgram struct {
    Id int `json:"id"`
    Station string `json:"directory_name"`
    Name OnsenProgramInfo `json:"program_info"`
    Contents []OnsenContent `json:"contents"`
}

func (prog *Program) String() string {
	return fmt.Sprintf("station: %s\ntitle: %s\ndate: %s\nurl: %s\n",
		prog.Station, prog.Title, prog.Date, prog.Url)
}

func GetStations() ([]string, error) {
	res, err := http.Get(ONSEN_PROGRAM_JSON)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

    type program struct {
        Station string `json:"directory_name"`
    }

    var v program
    var stations []string

    dec := json.NewDecoder(res.Body)
    // read open bracket
    if _, err := dec.Token(); err != nil {
        return nil, err
    }
    for dec.More(){
        if err := dec.Decode(&v); err != nil{
            return nil, err
        }
        append(stations, v.Station)
    }
    if _, err := dec.Token(); err != nil {
        return nil, err
    }

	sort.Strings(stations)

	return stations, nil
}

func GetOutputFilename(prog *Program, fileout string) (string, error) {
	s := fileout
	if s == "" {
		s = fmt.Sprintf("%s_%s_%s.mp3", prog.Title, prog.Id, prog.Date)
	}
	if s[0:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		s = strings.Replace(s, "~", usr.HomeDir, 1)
	}
	return strings.Replace(s, `/`, `_`, -1), nil
}

func Download(prog *Program, fileout string) error {
	res, err := http.Get(prog.Url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	fileout, err = GetOutputFilename(prog, fileout)
	if err != nil {
		return err
	}
	if _, err := os.Stat(fileout); err == nil {
		return fmt.Errorf(`File %s exists.`, fileout)
	}
	out, err := os.Create(fileout)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, res.Body)
	if err != nil {
		return err
	}
	return nil
}

func GetProgram(station string) (*Program, error) {
	res, err := http.Get(fmt.Sprintf("%s/%s", ONSEN_PROGRAM_JSON, station))
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
    
    var v OnsenProgram
    dec := json.NewDecoder(res.Body)
    if err := dec.Decode(&v); err != nil{
        return nil, err
    }

    prog := &Program{Station: v.Station, Title: v.Name}
    for content := range(v.Contents) {
        if conentent.Premium {
            continue
        }
        if content.Latest {
            prog.Url = content.Url
            prog.Id = content.Id
            prog.Date = strings.repalce(content.Date, "/", "_", -1)
        }
    }

	if len(mediaurl) == 0 {
		return prog, fmt.Errorf(station + ` exists, but no media found.`)
	}
	return prog, nil
}

func has(list []string, name string) bool {
	for _, v := range list {
		if name == v {
			return true
		}
	}
	return false
}
