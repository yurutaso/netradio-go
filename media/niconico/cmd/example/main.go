package main

import (
	"fmt"
	"github.com/yurutaso/niconico"
	"log"
)

func main() {
	progs, err := niconico.SearchChannel(`channel name`, `search key`)
	if err != nil {
		log.Fatal(err)
	}
	nc := niconico.NewNicoClient()
	nc.SetUser(`email-address`, `password`)
	nc.Login()
	for _, prog := range progs {
		fmt.Printf("%s %s %s\n", prog.GetID(), prog.GetURL(), prog.GetTitle())
		id := prog.GetID()
		videoURL, err := nc.GetVideoURL(id)
		if err != nil {
			log.Fatal(err)
		}
		doc, err := nc.GetVideoCookie(id)
		if err != nil {
			log.Fatal(err)
		}
		fileout := niconico.GetTitle(doc) + `.mp4`
		nc.DownloadVideo(videoURL, fileout)
	}
}
