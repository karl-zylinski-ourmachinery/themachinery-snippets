SNIPPETS

git stash && git pull --rebase && git push && git stash pop && tmbuild

---

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


* Mask indoors
* Mask outdoors for large groups
* 3ft distance
* Ventilation

If someone gets sick:

* Respond and report
* Quarantine and notify

* All the staff are vaccinated

* Will educate during Quarantine
* Will not test symptoms -- self-test

* Class starts at 8:20 -- zero hour
* Teachers will be there at 8:05
* 5-8 classes start at 9:15 (9:10 minimum)
* Van from Park & Rid 8:02
* Drop of Park & Ride 3:30, Library 3:40 -- closed on Monday

* Clubs begin Mon 13th
* Fri: Tabletop gaming club???

QUESTIONS

* When do clubs start?

