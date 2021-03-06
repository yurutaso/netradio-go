package agqr

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"
	"time"
)

func Download(fileout, duration string) error {
	usr, err := user.Current()
	if fileout == `` {
		t := time.Now()
		fileout = fmt.Sprintf("%4d%02d%02d%02d%02d_AGQR.flv", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute())
	}
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

	d, err := time.ParseDuration(duration)
	if err != nil {
		return err
	}

	//cmd := exec.Command(`rtmpdump`, `-r`, `rtmp://fms-base1.mitene.ad.jp/agqr/aandg22`, `--live`, `-o`, fileout)
	// 2019/09/07
	cmd := exec.Command(`rtmpdump`, `-r`, `rtmp://fms-base1.mitene.ad.jp/agqr/aandg1`, `--live`, `-o`, fileout)
	if err := cmd.Start(); err != nil {
		return err
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case <-time.After(d):
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
	case err := <-done:
		if err != nil {
			return err
		}
	}
	return nil
}
