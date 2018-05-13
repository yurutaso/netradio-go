package hibiki

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/user"
	"path"
	"sort"
	"strconv"
	"strings"
)

const (
	HIBIKI_API = `https://vcms-api.hibiki-radio.jp/api/v1`
)

type Program struct {
	station string
	url     string
	count   string
	date    string
	title   string
	person  string
}

func (prog *Program) String() string {
	return fmt.Sprintf("station: %s\ntitle: %s\ndate: %s\ncount: %s\ncast: %s\nurl: %s\n",
		prog.station, prog.title, prog.date, prog.count, prog.person, prog.url)
}

func GetProgram(station string) (*Program, error) {
	u, err := url.Parse(HIBIKI_API)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `programs`, station)

	// Get radio information as XML
	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// Extract information from XML
	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	if data[`episode`] == nil {
		return nil, fmt.Errorf("Warning. No program found with the media %s\n", station)
	}
	episode := data[`episode`].(map[string]interface{})

	video_id_tmp := episode[`video`].(map[string]interface{})[`id`]
	videoid := strconv.Itoa(int(video_id_tmp.(float64)))
	title := episode[`program_name`].(string)
	count := episode[`name`].(string)
	YMS_str := (strings.Split(episode[`updated_at`].(string), ` `))[0]
	YMS := strings.Split(YMS_str, `/`)
	date := fmt.Sprintf("%s%s%s", YMS[0], YMS[1], YMS[2])
	person := data[`cast`].(string)

	// Get radio url (needs videoid obtained from the above XML)
	u, err = url.Parse(HIBIKI_API)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `videos`, `play_check`)
	q := u.Query()
	q.Set(`video_id`, videoid)
	u.RawQuery = q.Encode()

	req, err = http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	res, err = client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var media map[string]string
	if err := json.NewDecoder(res.Body).Decode(&media); err != nil {
		return nil, err
	}
	mediaurl := media[`playlist_url`]

	return &Program{station: station, url: mediaurl, title: title, person: person, count: count, date: date}, nil
}

func GetStations() ([]string, error) {
	client := &http.Client{}
	u, err := url.Parse(HIBIKI_API)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `programs`)

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var v []interface{}
	if err := json.NewDecoder(res.Body).Decode(&v); err != nil {
		return nil, err
	}
	stations := make([]string, len(v), len(v))
	for i, f := range v {
		data := f.(map[string]interface{})
		stations[i] = data[`access_id`].(string)
	}
	sort.Strings(stations)
	return stations, nil
}

func Download(prog *Program, fileout string) error {
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
	return exec.Command("ffmpeg", "-y", "-i", prog.url, "-c", "copy", "-bsf:a", "aac_adtstoasc", fileout).Run()
}
