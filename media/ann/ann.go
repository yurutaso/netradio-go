package ann

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	//"path"
	"regexp"
	//"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	ANN_ROOT        = `https://i.allnightnippon.com`
	ANN_AUTH        = `https://i.allnightnippon.com/auth`
	ANN_AUTHAPI     = `https://i.allnightnippon.com/auth/fromApi`
	ANN_LOGIN       = `https://i-api.allnightnippon.com/auth/login`
	ANN_LOGOUT      = `https://i-api.allnightnippon.com/auth/logout`
	ANN_WUGRGR_HOME = `/pg/pg_anni_wugrgr`
	USER_AGENT      = `Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.189 Safari/537.36 Vivaldi/1.95.1077.55`
)

var rep_person = regexp.MustCompile(`[＜<][^＞>]*[＞>]`)

type Program struct {
	url    string
	m3u8   string
	title  string
	person string
}

func printBody(res *http.Response) {
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
}

func (prog *Program) String() string {
	return fmt.Sprintf("title: %s\ncast: %s\nurl: %s\n",
		prog.title, prog.person, prog.url)
}

func Login(client *http.Client, id string, password string) error {
	// set cookie jar
	jar, err := cookiejar.New(nil)
	if err != nil {
		return err
	}
	client.Jar = jar

	// login
	form := url.Values{
		`login_id`:   []string{id},
		`pw`:         []string{password},
		`auto_login`: []string{"0"},
		`api_token`:  []string{""},
	}
	req, err := http.NewRequest(`POST`, ANN_LOGIN, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set(`User-Agent`, USER_AGENT)
	req.Header.Set(`Host`, `i.allnightnippon.com`)
	req.Header.Set(`Origin`, `https://i.allnightnippon.com`)
	req.Header.Set(`Referer`, `https://i.allnightnippon.com/auth`)
	req.Header.Set(`Content-Type`, `application/x-www-form-urlencoded`)
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var v struct {
		Token  string `json:"api_token"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	if strings.ToUpper(v.Status) != `SUCCESS` {
		return fmt.Errorf("Failed to login. fromApi returns status code %s", v.Status)
	}

	// set the api_token to cookieJar
	u, err := url.Parse(ANN_AUTHAPI)
	if err != nil {
		return err
	}

	cookies := make([]*http.Cookie, 1)
	cookies[0] = &http.Cookie{Name: `api_token`, Value: v.Token}
	client.Jar.SetCookies(u, cookies)

	// validate api_token
	req, err = http.NewRequest(`POST`, ANN_AUTHAPI, nil)
	if err != nil {
		return err
	}
	req.Header.Set(`User-Agent`, USER_AGENT)
	//req.Header.Set(`Content-Type`, "application/x-www-form-urlencoded")
	//req.Header.Set(`Host`, `i.allnightnippon.com`)
	//req.Header.Set(`Origin`, `https://i.allnightnippon.com/auth`)
	//req.Header.Set(`Referer`, "https://i.allnightnippon.com/auth/login")
	//req.Header.Set(`Upgrade-Insecure-Requests`, "1")
	original := client.CheckRedirect
	client.CheckRedirect = func(req *http.Request, vis []*http.Request) error {
		return http.ErrUseLastResponse
	}

	res, err = client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	client.CheckRedirect = original

	fmt.Println(`Successfully login to anni`)
	return nil
}

func getM3U8(client *http.Client, u string) (m3u8 string, err error) {
	req, err := http.NewRequest(`GET`, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set(`User-Agent`, USER_AGENT)
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}

	m3u8, exists := doc.Find(`source`).Attr(`src`)
	if !exists {
		return "", fmt.Errorf("Cannot find mediaurl in %s", u)
	}
	return m3u8, nil
}

func GetLatestProgram(client *http.Client) ([]*Program, error) {
	req, err := http.NewRequest(`GET`, fmt.Sprintf("%s/pg/pg_anni_wugrgr?page=01", ANN_ROOT), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(`User-Agent`, USER_AGENT)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}

	progs := make([]*Program, 1)
	doc.Find(`div#ct_movie>div.inner>ul`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		prog := &Program{}

		// get url of the program page
		u, exists := s.Find(`a`).First().Attr(`href`)
		if !exists {
			err = fmt.Errorf("Cannot find a radio link in parsing page #01")
			return true
		}
		prog.url = fmt.Sprintf("%s/%s", ANN_ROOT, u)

		// get url of the m3u8 stream
		m3u8, err := getM3U8(client, prog.url)
		if err != nil {
			return true
		}
		prog.m3u8 = m3u8

		// get title
		prog.title = s.Find(`div.ttl_ct_program`).First().Text()
		if len(prog.title) == 0 {
			err = fmt.Errorf("Cannot find a title for the program %s", prog.url)
			return true
		}

		// extract personality from title
		if rep_person.MatchString(prog.title) {
			_person := rep_person.FindString(prog.title)
			_person = strings.Replace(_person, ">", "", -1)
			_person = strings.Replace(_person, "<", "", -1)
			_person = strings.Replace(_person, "＞", "", -1)
			_person = strings.Replace(_person, "＜", "", -1)
			prog.person = _person
		}
		progs[0] = prog
		return false
	})
	if err != nil {
		return nil, err
	}
	return progs, nil
}

func findProgramsInPage(page int, client *http.Client) ([]*Program, error) {
	req, err := http.NewRequest(`GET`, fmt.Sprintf("%s/pg/pg_anni_wugrgr?page=%d", ANN_ROOT, page), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(`User-Agent`, USER_AGENT)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}

	var m3u8 string
	progs := make([]*Program, 0)
	doc.Find(`div#ct_movie>div.inner>ul`).EachWithBreak(func(i int, s *goquery.Selection) bool {
		prog := &Program{}

		// get url of the program page
		u, exists := s.Find(`a`).First().Attr(`href`)
		if !exists {
			err = fmt.Errorf("Cannot find a radio link in parsing page #%d", page)
			return false
		}
		prog.url = fmt.Sprintf("%s/%s", ANN_ROOT, u)

		// get url of the m3u8 stream
		m3u8, err = getM3U8(client, prog.url)
		if err != nil {
			return false
		}
		prog.m3u8 = m3u8

		// get title
		prog.title = s.Find(`div.ttl_ct_program`).First().Text()
		if len(prog.title) == 0 {
			err = fmt.Errorf("Cannot find a title for the program %s", prog.url)
			return false
		}

		// extract personality from title
		if rep_person.MatchString(prog.title) {
			_person := rep_person.FindString(prog.title)
			_person = strings.Replace(_person, ">", "", -1)
			_person = strings.Replace(_person, "<", "", -1)
			_person = strings.Replace(_person, "＞", "", -1)
			_person = strings.Replace(_person, "＜", "", -1)
			prog.person = _person
		}
		progs = append(progs, prog)
		return true
	})
	if err != nil {
		return nil, err
	}
	return progs, nil
}

func GetPrograms(client *http.Client, max int) ([]*Program, error) {
	progs := make([]*Program, 0)

	// Loop through pages
	for page := 1; true; page++ {
		num := len(progs)
		_progs, err := findProgramsInPage(page, client)
		if err != nil {
			return nil, err
		}
		progs = append(progs, _progs...)

		// check whether to break
		n := len(progs)
		// break if no program is found
		if n == num {
			break
		}
		// break if the number of programs found exceeds the max
		if max > 0 && n >= max {
			progs = progs[0:max]
			break
		}
	}
	fmt.Printf("%d programs found.\n", len(progs))
	return progs, nil
}

func Download(client *http.Client, prog *Program, fileout string) error {
	if len(fileout) == 0 {
		fileout = fmt.Sprintf("%s.m4a", prog.title)
		fileout = strings.Replace(fileout, ">", "＞", -1)
		fileout = strings.Replace(fileout, "<", "＜", -1)
		fileout = strings.Replace(fileout, "!", "！", -1)
		fileout = strings.Replace(fileout, "?", "？", -1)
	}
	if _, err := os.Stat(fileout); err == nil {
		fmt.Printf("File %s exists.\n", fileout)
		return nil
	}
	fmt.Printf("Download file in %s as %s\n", prog.m3u8, fileout)
	err := exec.Command("ffmpeg", "-y", "-i", prog.m3u8, "-acodec", "copy", "-bsf:a", "aac_adtstoasc", fileout).Run()
	if err != nil {
		return err
	}
	return nil
}
