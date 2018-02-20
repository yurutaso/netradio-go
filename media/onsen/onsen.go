package onsen

import (
	"encoding/json"
	"fmt"
	"github.com/yurutaso/go-anirad"
	"io/ioutil"
	"net/http"
	"net/url"
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
	defer res.Body.Close()
	if err != nil {
		return err
	}
	if len(fileout) == 0 {
		fileout = anirad.ParseFilepath(prog.title + `_` + prog.count + `.m4a`)
	}
	return anirad.Download(res.Body, fileout)
}

func GetProgram(station string) (*Program, error) {
	stations, err := GetStations()
	if err != nil {
		return nil, err
	}
	if !anirad.Has(stations, station) {
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
	fmt.Println(data)
	if _, ok := data[`err`]; ok {
		return nil, fmt.Errorf(`Content named %v not found.`, station)
	}

	title := data[`title`].(string)
	count := data[`count`].(string)
	YMS := strings.Split(data[`update`].(string), `.`)
	date := fmt.Sprint("%s%s%s", YMS[0], YMS[1], YMS[2])
	person := data[`personality`].(string)
	mediaurl := ((data[`moviePath`].(map[string]interface{}))[`pc`]).(string)
	if len(mediaurl) == 0 {
		err := fmt.Errorf(station + ` exists, but no media found.`)
		return nil, err
	}
	return &Program{station: station, url: mediaurl, title: title, count: count, date: date, person: person}, nil
}
