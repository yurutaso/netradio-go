package agqr

import (
	"fmt"
	"os/exec"
	"os/user"
	"strings"
)

func (rad *agqr) Download(url, fileout string) error {
	// url is not used
	usr, err := user.Current()
	if err != nil {
		return err
	}
	fileout = strings.Replace(fileout, "~", usr.HomeDir, 1)

	if _, err := os.Stat(fileout); err == nil {
		return fmt.Errorf(`File %s exists.`, fileout)
	}

	if _, err := os.Stat(fileout); err == nil {
		return fmt.Errorf(`File %s exists`, fileout)
	}

	return exec.Command(`rtmpdump`, `-r`, `rtmp://fms-base1.mitene.ad.jp/agqr/aandg22`, `--live`, `-o`, parseFilepath(fileout)).Run()
}
