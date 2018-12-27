package radiko

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
)

var AREA_TABLE = map[string]string{
	`Hokkaido`:  `JP1`,
	`Aomori`:    `JP2`,
	`Iwate`:     `JP3`,
	`Miyagi`:    `JP4`,
	`Akita`:     `JP5`,
	`Yamagata`:  `JP6`,
	`Fukushima`: `JP7`,
	`Ibaragi`:   `JP8`,
	`Tochigi`:   `JP9`,
	`Gunma`:     `JP10`,
	`Saitama`:   `JP11`,
	`Chiba`:     `JP12`,
	`Tokyo`:     `JP13`,
	`Kanagawa`:  `JP14`,
	`Niigata`:   `JP15`,
	`Toyama`:    `JP16`,
	`Ishikawa`:  `JP17`,
	`Fukui`:     `JP18`,
	`Yamanashi`: `JP19`,
	`Nagano`:    `JP20`,
	`Gifu`:      `JP21`,
	`Shizuoka`:  `JP22`,
	`Aichi`:     `JP23`,
	`Mie`:       `JP24`,
	`Shiga`:     `JP25`,
	`Kyoto`:     `JP26`,
	`Osaka`:     `JP27`,
	`Hyogo`:     `JP28`,
	`Nara`:      `JP29`,
	`Wakayama`:  `JP30`,
	`Tottori`:   `JP31`,
	`Shimane`:   `JP32`,
	`Okayama`:   `JP33`,
	`Hiroshima`: `JP34`,
	`Yamaguchi`: `JP35`,
	`Tokushima`: `JP36`,
	`Kagawa`:    `JP37`,
	`Ehime`:     `JP38`,
	`Kochi`:     `JP39`,
	`Fukuoka`:   `JP40`,
	`Saga`:      `JP41`,
	`Nagasaki`:  `JP42`,
	`Kumamoto`:  `JP43`,
	`Oita`:      `JP44`,
	`Miyazaki`:  `JP45`,
	`Kagoshima`: `JP46`,
	`Okinawa`:   `JP47`,
}

type Program struct {
	station  string
	id       string
	ft       int64
	to       int64
	dur      int
	title    string
	info     string
	person   string
	videoURL string
}

func FilterByString(progs []Program, key string, fields ...string) ([]Program, error) {
	if fields == nil {
		fields = []string{`title`, `info`, `person`}
	}
	result := make([]Program, 0, 0)
	for _, prog := range progs {
		for _, field := range fields {
			data := ``
			switch field {
			case `title`:
				data = prog.title
			case `info`:
				data = prog.info
			case `person`:
				data = prog.person
			default:
				return nil, fmt.Errorf(`Unexpected field %s`, field)
			}
			if strings.Contains(data, key) {
				result = append(result, prog)
				break
			}
		}
	}
	return result, nil
}

func FilterByDate(progs []Program, from, to int64, field string) ([]Program, error) {
	if len(field) == 0 {
		field = `ft`
	}
	if field != `ft` && field != `to` {
		return nil, fmt.Errorf(`field must be "ft" or "to"`)
	}
	result := make([]Program, 0, 0)
	for _, prog := range progs {
		var data int64 = 0
		switch field {
		case `ft`:
			data = prog.ft
		case `to`:
			data = prog.to
		}
		if from < data && data < to {
			result = append(result, prog)
		}
	}
	return result, nil
}

func listAreaNames() {
	for key, val := range AREA_TABLE {
		fmt.Printf("areaName: %s, areaID: %s\n", key, val)
	}
}

func getAreaID(areaName string) (string, error) {
	if val, ok := AREA_TABLE[areaName]; ok {
		return val, nil
	}
	return "", fmt.Errorf("Invalid areaName. Run listAreaNames() to check available areaNames.")
}

func ListStations(areaID string) error {
	u, err := url.Parse(RADIKO_API_URL)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, `station`, `list`, areaID+`.xml`)
	res, err := http.Get(u.String())
	if err != nil {
		return err
	}
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return err
	}
	doc.Find("station").Each(func(_ int, s *goquery.Selection) {
		id := s.Find(`id`).First().Text()
		name := s.Find(`name`).First().Text()
		fmt.Printf("ID: %s, Name: %s\n", id, name)
	})
	return nil
}

func getAreaProgramXMLToday(areaID string) (io.ReadCloser, error) {
	u, err := url.Parse(RADIKO_API_URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `program`, `today`, areaID+`.xml`)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func getAreaProgramXMLNow(areaID string) (io.ReadCloser, error) {
	u, err := url.Parse(RADIKO_API_URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `program`, `now`, areaID+`.xml`)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func getAreaProgramXMLDate(areaID string, year, month, day int) (io.ReadCloser, error) {
	date := fmt.Sprintf("%4d%02d%02d", year, month, day)
	u, err := url.Parse(RADIKO_API_URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `program`, `date`, date, areaID+`.xml`)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func getStationProgramXMLToday(station string) (io.ReadCloser, error) {
	u, err := url.Parse(RADIKO_API_URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `program`, `station`, `today`, station+`.xml`)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func getStationProgramXMLWeek(station string) (io.ReadCloser, error) {
	u, err := url.Parse(RADIKO_API_URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `program`, `station`, `weekly`, station+`.xml`)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func getStationProgramXMLDate(station string, year, month, day int) (io.ReadCloser, error) {
	date := fmt.Sprintf("%4d%02d%02d", year, month, day)
	u, err := url.Parse(RADIKO_API_URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `program`, `station`, `date`, date, station+`.xml`)
	res, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	return res.Body, nil
}

func parseProgramXML(r io.Reader) ([]Program, error) {
	doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		return nil, err
	}
	progs := make([]Program, 0, 0)
	doc.Find(`station`).Each(func(_ int, s *goquery.Selection) {
		station, _ := s.Attr(`id`)
		doc.Find(`prog`).Each(func(_ int, s *goquery.Selection) {
			id, _ := s.Attr(`id`)
			_ft, _ := s.Attr(`ft`)
			//ft, err := strconv.Atoi(_ft)
			ft, err := strconv.ParseInt(_ft, 10, 64)
			if err != nil {
				log.Println(err)
			}
			_to, _ := s.Attr(`to`)
			//to, err := strconv.Atoi(_to)
			to, err := strconv.ParseInt(_to, 10, 64)
			if err != nil {
				log.Println(err)
			}
			_dur, _ := s.Attr(`dur`)
			dur, err := strconv.Atoi(_dur)
			//dur, err := strconv.ParseInt(_dur, 10, 64)
			if err != nil {
				log.Println(err)
			}
			person := s.Find(`pfm`).First().Text()
			info := s.Find(`info`).First().Text()
			title := s.Find(`title`).First().Text()
			progs = append(progs, Program{
				id:      id,
				ft:      ft,
				to:      to,
				dur:     dur,
				person:  person,
				info:    info,
				title:   title,
				station: station,
			})
		})
	})
	return progs, nil
}

func DateToInt(t time.Time) int64 {
	s := fmt.Sprintf("%4d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	//i, _ := strconv.Atoi(s)
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func GetAreaProgramToday(areaID string) ([]Program, error) {
	r, err := getAreaProgramXMLToday(areaID)
	if err != nil {
		return nil, err
	}
	return parseProgramXML(r)
}

func GetAreaProgramNow(areaID string) ([]Program, error) {
	r, err := getAreaProgramXMLNow(areaID)
	if err != nil {
		return nil, err
	}
	return parseProgramXML(r)
}

func GetAreaProgramDate(areaID string, year, month, day int) ([]Program, error) {
	r, err := getAreaProgramXMLDate(areaID, year, month, day)
	if err != nil {
		return nil, err
	}
	return parseProgramXML(r)
}

func GetStationProgramToday(station string) ([]Program, error) {
	r, err := getStationProgramXMLToday(station)
	if err != nil {
		return nil, err
	}
	return parseProgramXML(r)
}

func GetStationProgramWeek(station string) ([]Program, error) {
	r, err := getStationProgramXMLWeek(station)
	if err != nil {
		return nil, err
	}
	return parseProgramXML(r)
}

func GetStationProgramDate(station string, year, month, day int) ([]Program, error) {
	r, err := getStationProgramXMLDate(station, year, month, day)
	if err != nil {
		return nil, err
	}
	return parseProgramXML(r)
}

func Find(title, person, info, day, station string) ([]Program, error) {
	var progs []Program
	var err error
	if len(day) == 0 {
		progs, err = GetStationProgramWeek(station)
		if err != nil {
			return nil, err
		}
	} else {
		if len(day) != 8 {
			fmt.Println(day)
			return nil, fmt.Errorf("opt -day must be numbers YYYYMMDD (e.g. 20180527)")
		}
		y, err := strconv.Atoi(day[0:4])
		if err != nil {
			return nil, err
		}
		m, err := strconv.Atoi(day[4:6])
		if err != nil {
			return nil, err
		}
		d, err := strconv.Atoi(day[6:8])
		if err != nil {
			return nil, err
		}
		progs, err = GetStationProgramDate(station, y, m, d)
		if err != nil {
			return nil, err
		}
	}
	progs, err = FilterByString(progs, title, `title`)
	if err != nil {
		return nil, err
	}
	progs, err = FilterByString(progs, person, `person`)
	if err != nil {
		return nil, err
	}
	progs, err = FilterByString(progs, info, `info`)
	if err != nil {
		return nil, err
	}
	if len(progs) == 0 {
		return nil, fmt.Errorf(`No radio program found`)
	}
	return progs, nil
}
