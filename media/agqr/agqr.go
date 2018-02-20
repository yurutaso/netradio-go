package anirad

import (
	"fmt"
	"os/exec"
)

var (
	warnAGQRunsupported = fmt.Errorf(`AGQR only supports play and download`)
)

type agqr struct{}

func (rad *agqr) List() ([]string, error) {
	return nil, fmt.Errorf("agqr only supports play and download")
}

func (rad *agqr) GetProgram(name string) (*program, error) {
	return nil, fmt.Errorf("agqr only supports play and download")
}

func (rad *agqr) GetURL(name string) (string, error) {
	return "", fmt.Errorf("agqr only supports play and download")
}

func (rad *agqr) Download(url, fileout string) error {
	// url is not used
	fileout = parseFilepath(fileout)
	if exists(fileout) {
		return errFileExists
	}
	return exec.Command(`rtmpdump`, `-r`, `rtmp://fms-base1.mitene.ad.jp/agqr/aandg22`, `--live`, `-o`, parseFilepath(fileout)).Run()
}

func (rad *agqr) Stream(name string) error {
	// name is not used
	rtmpdump := exec.Command(`rtmpdump`, `-r`, `rtmp://fms-base1.mitene.ad.jp/agqr/aandg22`, `--live`)
	mpv := exec.Command(`mpv`, `-`, `-no-video`)
	return pipeCmd(rtmpdump, mpv)
}
