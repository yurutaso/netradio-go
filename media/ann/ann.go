package ann

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"os/exec"
	"path"
	"strings"
)

const (
	ANN_ROOT        = `https://i.allnightnippon.com`
	ANN_WUGRGR_HOME = `/pg/pg_anni_wugrgr`
)

type Program struct {
	url    string
	title  string
	count  string
	date   string
	person string
}

func (prog *Program) String() string {
	return fmt.Sprintf("title: %s\ndate: %s\ncount: %s\ncast: %s\nurl: %s\n",
		prog.title, prog.date, prog.count, prog.person, prog.url)
}

func GetProgram() (*Program, error) {
	client := &http.Client{}

	// Get the information of the latest program from /pg/pg_anni_wugrgr
	u, err := url.Parse(ANN_ROOT)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, ANN_WUGRGR_HOME)

	req, err := http.NewRequest(`GET`, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set(`User-Agent`, `Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.189 Safari/537.36 Vivaldi/1.95.1077.55`)
	res, err := client.Do(req)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}
	s := doc.Find(`div#container>div#ct_movie>div.inner>ul`).First()
	li1 := s.Find(`li`).First()
	link, exists := li1.Find(`a`).First().Attr(`href`)
	if !exists {
		return nil, fmt.Errorf(`Media not found.`)
	}

	li2 := s.Find(`li`).Next()
	info := li2.Find(`div.ttl_ct_program`).First().Text()

	substr := strings.Split(info, `】`)
	title := substr[1]
	substr = strings.Split(substr[0], `＜`)
	count := substr[0][6:]
	person := substr[1][:len(substr[1])-3]

	return &Program{url: link, title: title, count: count, person: person}, nil
}

func Download(prog *Program, fileout string) error {
	client := http.Client{}

	u, err := url.Parse(ANN_ROOT)
	if err != nil {
		return err
	}
	u.Path = path.Join(u.Path, prog.url)

	req, err := http.NewRequest(`GET`, u.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Set(`User-Agent`, `Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/64.0.3282.189 Safari/537.36 Vivaldi/1.95.1077.55`)
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return err
	}
	mediaurl, exists := doc.Find(`video.ulizahtml5>source`).Attr(`src`)
	if !exists {
		return fmt.Errorf(`m3u8 not found`)
	}

	if len(fileout) == 0 {
		fileout = prog.count + `_` + prog.title + `_` + prog.person + `.m4a`
	}
	err = exec.Command("ffmpeg", "-y", "-i", mediaurl, "-acodec", "copy", "-bsf:a", "aac_adtstoasc", fileout).Run()
	if err != nil {
		return err
	}
	return nil
}
