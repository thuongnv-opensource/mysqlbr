package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
	"time"
)

func ExecCommand(command *exec.Cmd) (error, string) {
	PrintBlue("[COMMAND]: ")
	fmt.Println(command.String())

	var outb, errb bytes.Buffer
	command.Stdout = &outb
	command.Stderr = &errb
	err := command.Run()
	if err != nil {
		return err, fmt.Sprintf("========== ERROR =========\n%s\n========== STDOUT =========\n%s\n========== STDERR =========\n%s", err.Error(), outb.String(), errb.String())
	}
	return nil, fmt.Sprintf("========== STDOUT =========\n%s\n========== STDERR =========\n%s", outb.String(), errb.String())
}

func ExecCommandStreamLog(command *exec.Cmd) (error, string) {
	PrintBlue("[COMMAND]: ")
	fmt.Println(command.String())

	stderr, _ := command.StderrPipe()
	stdout, _ := command.StdoutPipe()

	command.Start()

	var errStr string
	var outStr string

	go func() {
		scanner := bufio.NewScanner(stderr)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			errStr = errStr + m
			fmt.Println(m)
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdout)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			m := scanner.Text()
			fmt.Println(m)
			outStr = outStr + m
		}
	}()

	err := command.Wait()

	if err != nil {
		return err, fmt.Sprintf("========== ERROR =========\n%s\n========== STDOUT =========\n%s\n========== STDERR =========\n%s", err.Error(), outStr, errStr)
	}
	return nil, fmt.Sprintf("========== STDOUT =========\n%s\n========== STDERR =========\n%s", outStr, errStr)
}

func MakeTimestampFromNow() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func MakeTimestamp(t time.Time) int64 {
	result := t.UnixNano() / int64(time.Millisecond)
	return result
}

func MakeDirIfNotExist(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		os.MkdirAll(path, os.ModePerm)
	}
}

func CreateFileIfNotExist(path string) {
	var s = strings.Split(path, "/")
	var r = strings.Replace(path, "/"+s[len(s)-1], "", 1)
	MakeDirIfNotExist(r)
	var _, err = os.Stat(path)
	// create file if not exists
	if os.IsNotExist(err) {
		var file, err = os.Create(path)
		if err != nil {
			return
		}
		defer file.Close()
	}
}

type BackupRecord struct {
	Timestamp      *time.Time
	Base           *time.Time
	Incremental    *time.Time
	BaseStr        string
	IncrementalStr string
	Strategy       string
	HasError       bool
	ID             string
}

func GetBackupRecordPath(dataDir string, record BackupRecord) string {
	if strings.HasSuffix(dataDir, "/") {
		if record.Incremental == nil {
			return fmt.Sprintf("%s%s/%s/base", dataDir, record.Strategy, record.BaseStr)
		} else {
			return fmt.Sprintf("%s%s/%s/incremental/%s", dataDir, record.Strategy, record.BaseStr, record.IncrementalStr)
		}
	} else {
		if record.Incremental == nil {
			return fmt.Sprintf("%s/%s/%s/base", dataDir, record.Strategy, record.BaseStr)
		} else {
			return fmt.Sprintf("%s/%s/%s/incremental/%s", dataDir, record.Strategy, record.BaseStr, record.IncrementalStr)
		}
	}

}

func ParseBackupReportFile(path string) []BackupRecord {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Print(err)
	}
	str := string(b)
	result := make([]BackupRecord, 0)
	for _, s := range strings.Split(str, "\n") {
		if strings.TrimSpace(s) != "" {
			t := strings.Split(strings.TrimSpace(s), ":")
			if t[1] == "base" {
				p, _ := strconv.Atoi(t[2])
				var hasErr = false
				if t[3] == "error" {
					hasErr = true
				}
				baseT := time.UnixMilli(int64(p))
				result = append(result, BackupRecord{
					Strategy:    t[0],
					Timestamp:   &baseT,
					Base:        &baseT,
					Incremental: nil,
					HasError:    hasErr,
					BaseStr:     t[2],
					ID:          t[2],
				})
			} else if t[1] == "incremental" {
				p, _ := strconv.Atoi(t[2])
				baseT := time.UnixMilli(int64(p))
				p2, _ := strconv.Atoi(t[3])
				incrementalT := time.UnixMilli(int64(p2))
				var hasErr = false
				if t[4] == "error" {
					hasErr = true
				}

				result = append(result, BackupRecord{
					Strategy:       t[0],
					Timestamp:      &incrementalT,
					Base:           &baseT,
					Incremental:    &incrementalT,
					HasError:       hasErr,
					BaseStr:        t[2],
					IncrementalStr: t[3],
					ID:             t[3],
				})
			}
		}
	}
	sort.SliceStable(result, func(i, j int) bool {
		return result[i].Timestamp.UnixMilli() < result[j].Timestamp.UnixMilli()
	})
	return result
}

func AddBackupReportToFile(path string, pathToData string, strategy string, base int64, incremental int64, err error) {
	var strWrite string
	if incremental == 0 {
		strWrite = fmt.Sprintf("%s:base:%s", strategy, strconv.FormatInt(base, 10))
	} else {
		strWrite = fmt.Sprintf("%s:incremental:%s:%s", strategy, strconv.FormatInt(base, 10), strconv.FormatInt(incremental, 10))
	}
	if err != nil {
		strWrite = strWrite + ":error"
		PrintRed("[ADD BACKUP]")
		fmt.Println(": " + strWrite + "|" + pathToData)
	} else {
		strWrite = strWrite + ":success"
		PrintGreen("[ADD BACKUP]")
		fmt.Println(": " + strWrite + "|" + pathToData)
	}

	strWrite = strWrite + "\n"

	f, error := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if error != nil {
		panic(error)
	}
	defer f.Close()
	if _, error = f.WriteString(strWrite); error != nil {
		panic(error)
	}
}
