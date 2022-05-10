package service

import (
	"backup-mysql-cli/utils"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
)

func RsyncCopyFile(source string, destination string, sshPassword string) (error, string) {
	if destination == "" {
		dir, err := ioutil.TempDir("", "mysqlbr")
		if err != nil {
			log.Fatal(err)
		}
		destination = dir
	}
	s := strings.Split(source, "/")
	fileName := s[len(s)-1]

	c := exec.Command("rsync", source, destination)

	if sshPassword != "" {
		buffer := bytes.Buffer{}
		buffer.Write([]byte(sshPassword + "\n"))
		c.Stdin = &buffer
	}

	fmt.Println(c.String())
	err, _ := utils.ExecCommand(c)
	return err, destination + "/" + fileName
}

func RsyncCopyDir(source string, destination string, sshPassword string) (error, string) {
	if destination == "" {
		dir, err := ioutil.TempDir("", "mysqlbr")
		if err != nil {
			log.Fatal(err)
		}
		destination = dir
	}

	c := exec.Command("rsync", "-rz", source, destination)

	if sshPassword != "" {
		buffer := bytes.Buffer{}
		buffer.Write([]byte(sshPassword + "\n"))
		c.Stdin = &buffer
	}

	fmt.Println(c.String())
	err, _ := utils.ExecCommand(c)
	return err, destination
}
