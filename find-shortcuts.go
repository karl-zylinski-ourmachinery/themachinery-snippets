// Find all keyboard checks in the code so that we can convert them to shortcuts.

// TODO: Check modifiers
// TODO: Check edit keys

package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"path"
	"path/filepath"
	"regexp"
	"strings"
)

// This are considered standard editing keys that it doesn't make sense to remap.
var ignoreKeys = map[string]bool{
	"TM_INPUT_KEYBOARD_ITEM_1":           true,
	"TM_INPUT_KEYBOARD_ITEM_2":           true,
	"TM_INPUT_KEYBOARD_ITEM_3":           true,
	"TM_INPUT_KEYBOARD_ITEM_4":           true,
	"TM_INPUT_KEYBOARD_ITEM_5":           true,
	"TM_INPUT_KEYBOARD_ITEM_6":           true,
	"TM_INPUT_KEYBOARD_ITEM_7":           true,
	"TM_INPUT_KEYBOARD_ITEM_8":           true,
	"TM_INPUT_KEYBOARD_ITEM_9":           true,
	"TM_INPUT_KEYBOARD_ITEM_0":           true,
	"TM_INPUT_KEYBOARD_ITEM_LEFT":        true,
	"TM_INPUT_KEYBOARD_ITEM_RIGHT":       true,
	"TM_INPUT_KEYBOARD_ITEM_UP":          true,
	"TM_INPUT_KEYBOARD_ITEM_DOWN":        true,
	"TM_INPUT_KEYBOARD_ITEM_ENTER":       true,
	"TM_INPUT_KEYBOARD_ITEM_SPACE":       true,
	"TM_INPUT_KEYBOARD_ITEM_ESCAPE":      true,
	"TM_INPUT_KEYBOARD_ITEM_COUNT":       true,
	"TM_INPUT_KEYBOARD_ITEM_NONE":        true,
	"TM_INPUT_KEYBOARD_ITEM_END":         true,
	"TM_INPUT_KEYBOARD_ITEM_HOME":        true,
	"TM_INPUT_KEYBOARD_ITEM_BACKSPACE":   true,
	"TM_INPUT_KEYBOARD_ITEM_DELETE":      true,
	"TM_INPUT_KEYBOARD_ITEM_PAGEUP":      true,
	"TM_INPUT_KEYBOARD_ITEM_PAGEDOWN":    true,
	"TM_INPUT_KEYBOARD_ITEM_NUMPADENTER": true,
	"TM_INPUT_KEYBOARD_ITEM_LEFTALT":     true,
	"TM_INPUT_KEYBOARD_ITEM_NUMPADPLUS":  true,
	"TM_INPUT_KEYBOARD_ITEM_NUMPADMINUS": true,
}

var ignoreFiles = map[string]bool{
	`input.osx.c`:        true,
	`input.linux.c`:      true,
	`input.c`:            true,
	`ui.c`:               true,
	`shortcut_manager.c`: true,
	`ui_controls.c`:      true,
	`sculpt_component.c`: true,

	`modal.c`:        true,
	`scene_common.c`: true,
}

var ignoreDirs = map[string]bool{
	`epsilon-machine`: true,
	`.git`:            true,
	`simple_3d`:       true,
	`simple_draw`:     true,
}

func findKeyboardItems(s, file string) {
	if ignoreFiles[path.Base(file)] {
		return
	}
	re := regexp.MustCompile(`TM_INPUT_KEYBOARD_ITEM_[\w]*`)
	locs := re.FindAllStringIndex(s, -1)
	for _, loc := range locs {
		key := s[loc[0]:loc[1]]
		if ignoreKeys[key] {
			continue
		}
		line := strings.Count(s[:loc[0]], "\n") + 1
		fmt.Printf("%s:%d %s\n", file, line, key)
	}
}

func processString(s, file string) {
	findKeyboardItems(s, file)
}

func processFile(file string) {
	dat, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	s := string(dat)
	processString(s, file)
}

func walkDirFunc(file string, d fs.DirEntry, err error) error {
	if ignoreDirs[path.Base(file)] {
		return fs.SkipDir
	} else if !d.IsDir() && strings.HasSuffix(d.Name(), ".c") {
		processFile(file)
	}
	return nil
}

func main() {
	filepath.WalkDir("..", walkDirFunc)
}
