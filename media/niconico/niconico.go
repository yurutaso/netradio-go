package niconico

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

const (
	loginURL     = `https://secure.nicovideo.jp/secure/login`
	getflvURL    = `http://flapi.nicovideo.jp/api/getflv/`
	watchURL     = `http://www.nicovideo.jp/watch/`
	liveinfoURL  = `http://watch.live.nicovideo.jp/api/getplayerstatus/lv`
	livewatchURL = `http://live.nicovideo.jp/watch/lv`
)

type NicoClient struct {
	jar      *cookiejar.Jar
	client   *http.Client
	password string
	email    string
}

type LiveVideo struct {
	videoURLs []string
	ticket    string
	rtmpurl   string
}

func NewNicoClient() *NicoClient {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{Jar: jar}
	return &NicoClient{jar: jar, client: client}
}

func (nc *NicoClient) SetUser(email, password string) {
	nc.password = password
	nc.email = email
}

func (nc *NicoClient) Login() error {
	values := url.Values{`mail_tel`: []string{nc.email}, `password`: []string{nc.password}}
	res, err := nc.client.PostForm(loginURL, values)
	defer res.Body.Close()
	if err != nil {
		return err
	}
	return nil
}

func (nc *NicoClient) GetVideoURL(videoID string) (string, error) {
	res, err := nc.client.Get(getflvURL + videoID)
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
	return values[`url`][0], nil
}

func (nc *NicoClient) GetVideoCookie(videoID string) (*goquery.Document, error) {
	res, err := nc.client.Get(watchURL + videoID)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func GetTitle(doc *goquery.Document) string {
	title := doc.Find(`title`).Text()
	title = strings.Replace(title, `- ニコニコ動画:GINZA`, "", -1)
	title = strings.Replace(title, `- ニコニコ動画`, "", -1)
	title = strings.Replace(title, `/`, "_", -1)
	return strings.TrimSpace(title)
}

func (nc *NicoClient) GetLiveInfo(liveID string) (*LiveVideo, error) {
	res, err := nc.client.Get(liveinfoURL + liveID)
	//defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	rtmpurl := doc.Find(`url`).Text()
	ticket := doc.Find(`ticket`).Text()

	que := doc.Find(`que`).Text()
	re := regexp.MustCompile(`(\/content\/[0-9]*\/lv[^.]*\.f4v)\/`)
	result := re.FindAllStringSubmatch(que, -1)
	videos := make([]string, len(result), len(result))
	for i, found := range result {
		videos[i] = found[1]
	}
	return &LiveVideo{videoURLs: videos, ticket: ticket, rtmpurl: rtmpurl}, nil
}

func (nc *NicoClient) GetLiveTitle(liveID string) (string, error) {
	res, err := nc.client.Get(livewatchURL + liveID)
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}
	title := ``
	doc.Find(`h1.title_text`).Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})
	return title, nil
}

func GetFileBaseExt(fileout string) (string, string, error) {
	ext := filepath.Ext(fileout)
	fileout_base := fileout[:len(fileout)-len(ext)]
	return fileout_base, ext, nil
}

func (nc *NicoClient) DownloadTimeshift(liveVideo *LiveVideo, fileout string) {
	rtmpurl := liveVideo.rtmpurl
	ticket := liveVideo.ticket
	base, ext, err := GetFileBaseExt(fileout)
	if err != nil {
		log.Fatal(err)
	}
	for i, video := range liveVideo.videoURLs {
		if i > 0 {
			fileout = base + `_` + strconv.Itoa(i) + ext
		} else {
			fileout = base + ext
		}
		fmt.Println(`Saving video as ` + fileout)
		cmd := exec.Command(`rtmpdump`, `-r`, rtmpurl, `-y`, `mp4:`+video, `-C`, `S:`+ticket, `-o`, fileout)
		//fmt.Println(cmd)
		result, err := cmd.CombinedOutput()
		fmt.Println(string(result))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func (nc *NicoClient) DownloadVideo(videoURL, fileout string) {
	fmt.Println(`Saving video as ` + fileout)
	out, err := os.Create(fileout)
	if err != nil {
		log.Fatal(err)
	}
	defer out.Close()
	res, err := nc.client.Get(videoURL)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	_, err = io.Copy(out, res.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(`Complete!`)
}
