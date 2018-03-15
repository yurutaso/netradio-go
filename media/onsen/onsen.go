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
	Station string
	Url     string
	Title   string
	Count   string
	Date    string
	Person  string
}

func (prog *Program) String() string {
	return fmt.Sprintf("station: %s\ntitle: %s\ndate: %s\ncount: %s\ncast: %s\nurl: %s\n",
		prog.Station, prog.Title, prog.Date, prog.Count, prog.Person, prog.Url)
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
	res, err := http.Get(prog.Url)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if len(fileout) == 0 {
		fileout = prog.Title + `_` + prog.Count + `.mp3`
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
	if data == nil {
		return nil, fmt.Errorf(`No media found`)
	}

	// Check if API returns nothing or error json.
	//fmt.Println(data)
	if _, ok := data[`err`]; ok {
		return nil, fmt.Errorf(`Content named %v not found.`, station)
	}

	title := ``
	count := ``
	date := ``
	person := ``
	mediaurl := ``
	if data[`title`] != nil {
		title = data[`title`].(string)
	}
	if data[`count`] != nil {
		count = data[`count`].(string)
	}
	if data[`update`] != nil {
		YMS := strings.Split(data[`update`].(string), `.`)
		if len(YMS) > 2 {
			date = fmt.Sprintf("%s%s%s", YMS[0], YMS[1], YMS[2])
		}
	}
	if data[`personality`] != nil {
		person = data[`personality`].(string)
	}
	if data[`moviePath`] != nil {
		pathes := data[`moviePath`].(map[string]interface{})
		if pathes[`pc`] != nil {
			mediaurl = pathes[`pc`].(string)
		}
	}
	prog := &Program{Station: station, Url: mediaurl, Title: title, Count: count, Date: date, Person: person}
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
