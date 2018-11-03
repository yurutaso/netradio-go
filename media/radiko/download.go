package radiko

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"time"
)

func GetOutputFilename(prog Program, fileout string) (string, error) {
	s := fileout
	if s == "" {
		s = strconv.Itoa(prog.ft) + `_` + prog.title + `.mp3`
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

func Download(fileout string, prog Program) {
	// exit if future program
	if t := DateToInt(time.Now()); t < prog.to {
		log.Fatal(fmt.Sprintf("Latest program found on %s.\nThe last program might be canceled by baseball game or any other reason.", strconv.Itoa(prog.to)))
	}
	//client, token, err := login()
	client, token, err := login2()
	if err != nil {
		log.Fatal(err)
	}

	m3u8, err := getM3U8(client, token, prog.station, prog.ft, prog.to)
	if err != nil {
		log.Fatal(err)
	}

	fileout, err = GetOutputFilename(prog, fileout)
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(
		`ffmpeg`,
		`-headers`, `"X-Radiko-AuthToken: `+token+`"`,
		`-i`, m3u8,
		`-vn`,
		fileout)
	fmt.Println(`%s `, cmd)
	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", b)
}

func getM3U8(client *http.Client, token, station_id string, ft, to int) (string, error) {
	ft_i := strconv.Itoa(ft)
	to_i := strconv.Itoa(to)

	values := url.Values{}
	values.Set(`station_id`, station_id)
	values.Set(`l`, `15`)
	values.Set(`ft`, ft_i)
	values.Set(`to`, to_i)
	u, err := url.Parse(RADIKO_PLAYLIST_URL)
	if err != nil {
		return "", err
	}
	u.RawQuery = values.Encode()

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Add(`X-Radiko-AuthToken`, token)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	text := string(b)
	if text == `expired` {
		return "", fmt.Errorf("This program is expired. ")
	}
	m3u8 := strings.Split(text, "\n")[3]
	return m3u8, err
}
