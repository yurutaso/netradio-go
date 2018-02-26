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
	"sort"
	"strings"
)

const (
	ONSEN_PROGRAM_JSON = `http://www.onsen.ag/api/shownMovie/shownMovie.json`
	ONSEN_PROGRAM_API  = `http://www.onsen.ag/data/api`
)

type Program struct {
	station string
	url     string
	title   string
	count   string
	date    string
	person  string
}

func (prog *Program) String() string {
	return fmt.Sprintf("station: %s\ntitle: %s\ndate: %s\ncount: %s\ncast: %s\nurl: %s\n",
		prog.station, prog.title, prog.date, prog.count, prog.person, prog.url)
}

func GetStations() ([]string, error) {
	res, err := http.Get(ONSEN_PROGRAM_JSON)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	var v map[string][]string
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, err
	}
	if _, ok := v[`result`]; !ok {
		return nil, fmt.Errorf(`key "result" does not exist in the onsen json.`)
	}
	result := v[`result`]
	sort.Strings(result)
	return result, nil
}

func Download(prog *Program, fileout string) error {
	res, err := http.Get(prog.url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if len(fileout) == 0 {
		fileout = prog.title + `_` + prog.count + `.m4a`
		// Sometimes prog.title contains `/`, which may cause error in creating new file
		fileout = strings.Replace(fileout, `/`, `_`, -1)
	}
	usr, err := user.Current()
	if err != nil {
		return err
	}
	fileout = strings.Replace(fileout, "~", usr.HomeDir, 1)
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
	stations, err := GetStations()
	if err != nil {
		return nil, err
	}
	if !has(stations, station) {
		return nil, fmt.Errorf(`Station "%s" not found`, station)
	}

	// Set request to ONSEN API
	u, err := url.Parse(ONSEN_PROGRAM_API)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `getMovieInfo`, station)
	q := u.Query()
	q.Set("callback", "callback")
	u.RawQuery = q.Encode()

	// Get response from API
	res, err := http.Get(u.String())
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	// Cut prefix and suffix characters (prefix: `callback(`, suffix: `);`)
	text := string(b)
	text = text[9 : len(text)-3]

	// parse json
	var data map[string]interface{}
	err = json.Unmarshal([]byte(text), &data)
	if err != nil {
		return nil, err
	}

	// Check if API returns nothing or error json.
	//fmt.Println(data)
	if _, ok := data[`err`]; ok {
		return nil, fmt.Errorf(`Content named %v not found.`, station)
	}

	title := data[`title`].(string)
	count := data[`count`].(string)
	YMS := strings.Split(data[`update`].(string), `.`)
	date := fmt.Sprintf("%s%s%s", YMS[0], YMS[1], YMS[2])
	person := data[`personality`].(string)
	mediaurl := ((data[`moviePath`].(map[string]interface{}))[`pc`]).(string)
	if len(mediaurl) == 0 {
		err := fmt.Errorf(station + ` exists, but no media found.`)
		return nil, err
	}
	return &Program{station: station, url: mediaurl, title: title, count: count, date: date, person: person}, nil
}

func has(list []string, name string) bool {
	for _, v := range list {
		if name == v {
			return true
		}
	}
	return false
}
