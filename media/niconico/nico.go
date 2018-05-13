package anirad

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os/exec"
	"strings"
)

const (
	nicologin  = `https://secure.nicovideo.jp/secure/login`
	nicoapiflv = `http://flapi.nicovideo.jp/api/getflv/`
	nicowatch  = `http://www.nicovideo.jp/watch/`
)

var (
	//channel     = []string{`seaside-channel, anitama-ch`}
	errNotLogin = fmt.Errorf(`Not login`)
)

type Nico struct {
	jar    *cookiejar.Jar
	client *http.Client
	login  bool
}

func NewNico() *Nico {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	return &Nico{jar: jar, client: client, login: false}
}

func (nc *Nico) List() ([]string, error) {
	return []string{`nico does not support list command`}, nil
}
func (nc *Nico) GetProgram(videoID string) (*program, error) {
	res, err := nc.client.Get(nicowatch + videoID)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}

	title := doc.Find(`title`).Text()
	title = strings.Replace(title, `- ニコニコ動画:GINZA`, "", -1)
	title = strings.Replace(title, `- ニコニコ動画`, "", -1)
	title = strings.Replace(title, `/`, "_", -1)
	title = strings.TrimSpace(title)

	u := ""
	if u, err = nc.getURL(videoID); err != nil {
		if err != errNotLogin {
			return nil, err
		}
		fmt.Println(err.Error())
	}
	return &program{id: videoID, title: title, url: u}, nil
}

func (nc *Nico) Download(videoURL, fileout string) error {
	fileout = parseFilepath(fileout) + `.m4a`
	if exists(fileout) {
		return errFileExists
	}

	if !nc.login {
		return errNotLogin
	}

	res, err := nc.client.Get(videoURL)
	defer res.Body.Close()
	if err != nil {
		return err
	}

	ffmpeg := exec.Command("ffmpeg", "-y", "-i", "pipe:0", "-vn", "-acodec", "copy", fileout)
	pipeFromReader(res.Body, ffmpeg)
	return ffmpeg.Run()
}

func (nc *Nico) Stream(videoURL string) error {
	if !nc.login {
		return errNotLogin
	}

	res, err := nc.client.Get(videoURL)
	defer res.Body.Close()
	if err != nil {
		return err
	}

	mpv := exec.Command("mpv", "-", "--no-video")
	pipeFromReader(res.Body, mpv)
	return mpv.Run()
}

func (nc *Nico) Login(email, password string) error {
	values := url.Values{`mail_tel`: []string{email}, `password`: []string{password}}
	res, err := nc.client.PostForm(nicologin, values)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	nc.login = true
	return nil
}

func (nc *Nico) getVideoCookie(videoID string) error {
	if !nc.login {
		return errNotLogin
	}
	res, err := nc.client.Get(nicowatch + videoID)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func (nc *Nico) getURL(videoID string) (string, error) {
	if !nc.login {
		return "", errNotLogin
	}
	res, err := nc.client.Get(nicoapiflv + videoID)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}
	query := doc.Find(`Body`).Text()
	values, err := url.ParseQuery(query)
	if err != nil {
		return "", err
	}
	if len(values[`url`]) > 0 {
		nc.getVideoCookie(videoID)
		return values[`url`][0], nil
	}
	return "", fmt.Errorf(`Invalid videoID`)
}
