package hibiki

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os/exec"
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
	count   string
	date    string
	title   string
	person  string
}

func GetProgram(station string) (*Program, error) {
	u, err := url.Parse(HIBIKI_API)
	if err != nil {
		return nil
	}
	u.Path = path.Join(u.Path, `programs`, station)

	client := &http.Client{}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	res, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer res.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return nil, err
	}
	return m, nil
	if err != nil {
		return nil, err
	}

	if data[`episode`] == nil {
		return nil, nil
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

	client := &http.Client{}
	u, err := url.Parse(HIBIKI_API)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, `videos`, `play_check`)
	q := u.Query()
	q.Set(`video_id`, videoid)
	u.RawQuery = q.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add("X-Requested-With", "XMLHttpRequest")
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	var m map[string]string
	if err := json.NewDecoder(res.Body).Decode(&m); err != nil {
		return "", err
	}
	mediaurl := m[`playlist_url`]
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
	sort.Strings(list)
	return stations, nil
}

func Download(url, fileout string) error {
	fileout = parseFilepath(fileout)
	if exists(fileout) {
		return fmt.Errorf(`File %s exists.`, fileout)
	}
	return exec.Command("ffmpeg", "-y", "-i", url, "-c", "copy", "-bsf:a", "aac_adtstoasc", fileout).Run()
}
