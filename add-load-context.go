package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

const template = `void load_%s(struct tm_api_registry_api *reg, bool load)
{
    reg->begin_context("%s");
%s

    reg->end_context("%s");
}`

func processString(s string) string {
	re := regexp.MustCompile(`(?s)void load_(\w*)\(struct tm_api_registry_api \*reg, bool load\)\n\{(.*)\n\}`)

	// Dry run
	// res := re.FindAllString(s, -1)
	// for _, r := range res {
	// 	sub := re.FindStringSubmatch(r)
	// 	fmt.Printf(template, sub[1], sub[1], sub[2], sub[1])
	// }

	s = re.ReplaceAllStringFunc(s, func(match string) string {
		sub := re.FindStringSubmatch(match)
		return fmt.Sprintf(template, sub[1], sub[1], sub[2], sub[1])
	})

	return s
}

func processFile(path string) {
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
	} else if !d.IsDir() && strings.HasSuffix(d.Name(), ".c") {
		processFile(path)
	}
	return nil
}

func main() {
	filepath.WalkDir("..", walkDirFunc)
}
