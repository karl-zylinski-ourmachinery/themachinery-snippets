package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"strings"
)

func addSuperToTabVt(s string) string {
	re := regexp.MustCompile(`(?s)\(tm_the_machinery_tab_vt\)\{(.*?)\}`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		text := re.FindStringSubmatch(match)[1]
		return fmt.Sprintf("(tm_the_machinery_tab_vt){\n.super = {\n%s}\n}", text)
	})
	return s

}

func fixTabImplementation(s string) string {
	re := regexp.MustCompile(`tm_add_or_remove_implementation\(reg, load, tm_tab_vt, ([a-z0-9_]*)\);`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		text := re.FindStringSubmatch(match)[1]
		return fmt.Sprintf("tm_add_or_remove_implementation(reg, load, tm_tab_vt, &%s->super);", text)
	})
	return s
}

func processString(s string) string {
	return fixTabImplementation(s)
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
	filepath.WalkDir("../the_machinery", walkDirFunc)
}
