package niconico

import (
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

const (
	nicochannel = "http://ch.nicovideo.jp/"
)

type program struct {
	id    string
	url   string
	title string
	count string
	cast  string
	date  [3]string
}

func (prog *program) GetID() string {
	return prog.id
}
func (prog *program) GetURL() string {
	return prog.url
}
func (prog *program) GetTitle() string {
	return prog.title
}
func (prog *program) GetCount() string {
	return prog.count
}
func (prog *program) GetCast() string {
	return prog.cast
}
func (prog *program) GetDate() [3]string {
	return prog.date
}

func GetChannelID(channel string) (string, error) {
	res, err := http.Get(nicochannel + channel)
	defer res.Body.Close()
	if err != nil {
		return "", err
	}
	doc, err := goquery.NewDocumentFromResponse(res)
	if err != nil {
		return "", err
	}
	s, _ := doc.Find("a.thumb_ch").First().Attr("href")
	return s[1:], nil
}

func SearchChannel(channel string, keyword string) ([]*program, error) {
	channelID, err := GetChannelID(channel)
	if err != nil {
		return nil, err
	}
	u, err := url.Parse(`http://ch.nicovideo.jp/search/` + keyword)
	if err != nil {
		return nil, err
	}
	data := url.Values{}
	data.Add(`channel_id`, channelID)
	data.Add(`mode`, `s`)
	data.Add(`sort`, `f`)
	data.Add(`order`, `d`)
	data.Add(`type`, `video`)
	page := 1
	progs := make([]*program, 0, 50)

	for {
		data.Set(`page`, strconv.Itoa(page))
		u.RawQuery = data.Encode()
		res, err := http.Get(u.String())
		defer res.Body.Close()
		if err != nil {
			return nil, err
		}
		doc, err := goquery.NewDocumentFromResponse(res)
		if err != nil {
			return nil, err
		}
		l0 := len(progs)
		doc.Find(`div.item_right`).Each(func(i int, s *goquery.Selection) {
			data := s.Find("a")
			u, exists1 := data.Attr("href")
			title, exists2 := data.Attr("title")
			if exists1 && exists2 {
				videoID := path.Base(u)
				progs = append(progs, &program{title: title, id: videoID})
			}
		})
		if l0 == len(progs) {
			break
		}
		page++
	}
	return progs, nil
}
