package anirad

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strings"
)

func Has(list []string, name string) bool {
	for _, v := range list {
		if name == v {
			return true
		}
	}
	return false
}

func Exists(filepath string) bool {
	_, err := os.Stat(filepath)
	return err == nil
}

func PipeFromReader(reader io.Reader, cmd *exec.Cmd) error {
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		io.Copy(stdin, reader)
	}()
	return nil
}

func PipeCmd(cmdin, cmdout *exec.Cmd) error {
	stdin, err := cmdin.StdoutPipe()
	if err != nil {
		return err
	}
	cmdout.Stdin = stdin
	go func() {
		defer stdin.Close()
		cmdin.Start()
	}()

	return cmdout.Run()
}

func Download(writer io.Reader, fileout string) error {
	fileout = ParseFilepath(fileout)
	if Exists(fileout) {
		return fmt.Errorf(`File %s exists`, fileout)
	}

	out, err := os.Create(fileout)
	defer out.Close()
	if err != nil {
		return err
	}
	_, err = io.Copy(out, writer)
	if err != nil {
		return err
	}
	return nil
}

func ParseFilepath(fileout string) string {
	usr, _ := user.Current()
	fileout = strings.Replace(fileout, "~", usr.HomeDir, 1)
	return fileout
}
