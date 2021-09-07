package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
)

func main() {
	dat, err := ioutil.ReadFile("ui.c")
	if err != nil {
		panic(err)
	}
	s := string(dat)
	re := regexp.MustCompile(`\{ \.r = (\d+), \.g = (\d+), \.b = (\d+), \.a = 255 \}`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		sub := re.FindStringSubmatch(match)
		r, _ := strconv.Atoi(sub[1])
		g, _ := strconv.Atoi(sub[2])
		b, _ := strconv.Atoi(sub[3])
		return fmt.Sprintf("COLOR(0x%02x%02x%02x)", r, g, b)
	})
	re = regexp.MustCompile(`\{ (\d+), (\d+), (\d+), 255 \}`)
	s = re.ReplaceAllStringFunc(s, func(match string) string {
		sub := re.FindStringSubmatch(match)
		r, _ := strconv.Atoi(sub[1])
		g, _ := strconv.Atoi(sub[2])
		b, _ := strconv.Atoi(sub[3])
		return fmt.Sprintf("COLOR(0x%02x%02x%02x)", r, g, b)
	})
	ioutil.WriteFile("ui.c", []byte(s), 0644)
}
