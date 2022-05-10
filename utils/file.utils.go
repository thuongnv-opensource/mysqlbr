package utils

import (
	"io/ioutil"
	"strings"
)

func ReadFileInCurrentDirectory(fileName string) string {
	data, error := ioutil.ReadFile(fileName)
	if error != nil {
		panic(error)
	}
	return string(data)
}

func ResolveConfigPathFromDataDir(dataDir string) string {
	if strings.HasSuffix(dataDir, "/") {
		return dataDir + "bk.dat"
	} else {
		return dataDir + "/bk.dat"
	}
}
