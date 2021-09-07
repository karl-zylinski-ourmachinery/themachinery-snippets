package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func processString(s string) string {
	re := regexp.MustCompile(`(?s)struct (tm_[a-z0-9_]*_api)\n{.*?\n};`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		api := re.FindStringSubmatch(match)[1]
		return match + fmt.Sprintf("\n\n#define %s_version TM_VERSION(1, 0, 0)", api)
	})
	return s
}

func processHeader(path string) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	s := string(dat)
	s = processString(s)
	ioutil.WriteFile(path, []byte(s), 0644)
}

func walkDirFunc(path string, d fs.DirEntry, err error) error {
	if strings.HasSuffix(d.Name(), ".git") {
		return fs.SkipDir
	} else if !d.IsDir() && strings.HasSuffix(d.Name(), ".h") {
		processHeader(path)
	}
	return nil
}

func main() {
	filepath.WalkDir("..", walkDirFunc)
}
