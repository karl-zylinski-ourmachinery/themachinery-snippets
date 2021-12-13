package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

const settingsFile = "releaseBuild.json"

const STEP_CHECK_OUT_SOURCE = "Check out source code"
const STEP_CLEAN = "Clean directory"
const STEP_BUILD_WINDOWS_PACKAGE = "Build Windows package"
const STEP_TEST_WINDOWS_PACKAGE = "Test Windows package"
const STEP_UPDATE_VERSION_NUMBERS = "Update version numbers"
const STEP_COMMIT_CHANGES = "Commit changes"

// GetSetting returns the setting for the specified key.
func GetSetting(key string) string {
	data := make(map[string]string)
	bytes, err := ioutil.ReadFile(settingsFile)
	if err != nil {
		return ""
	}
	json.Unmarshal(bytes, &data)
	return data[key]
}

// SetSetting sets the setting for the specified key.
func SetSetting(key, value string) {
	data := make(map[string]string)
	bytes, err := ioutil.ReadFile(settingsFile)
	if err == nil {
		json.Unmarshal(bytes, &data)
	}
	data[key] = value
	txt, _ := json.MarshalIndent(data, "", "    ")
	ioutil.WriteFile(settingsFile, txt, 0644)
}

// If a setting exists for the specified prompt, returns that setting. Otherwise, prints the
// prompt and asks the user to type in the setting.
func ReadSetting(prompt string) string {
	s := GetSetting(prompt)
	if s != "" {
		return s
	}
	fmt.Print(prompt + ": ")
	fmt.Scanln(&s)
	SetSetting(prompt, s)
	return s
}

// Marks the step as completed for future runs of the program.
func CompleteStep(step string) {
	SetSetting(step, "true")
}

// Returns true if the step has been completed in a previous run of the program.
func HasCompletedStep(step string) bool {
	res := GetSetting(step) == "true"
	if !res {
		fmt.Println()
		fmt.Println("-------------------------------------------------------")
		fmt.Println(step)
		fmt.Println("-------------------------------------------------------")
		fmt.Println()
	}
	return res
}

// Runs the command, printing output and stopping execution in case of an error.
func Run(cmd *exec.Cmd) {
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func Instruction(s string) {
	fmt.Println(s)
	fmt.Println()
	fmt.Println("Press enter to continue...")
	fmt.Scanln()
}

func buildWindowsPackage() {
	if !HasCompletedStep(STEP_CLEAN) {
		Run(exec.Command("tmbuild", "--clean"))
		CompleteStep(STEP_CLEAN)
	}

	if !HasCompletedStep(STEP_BUILD_WINDOWS_PACKAGE) {
		Run(exec.Command("tmbuild", "-p", "release-package.json"))
		Run(exec.Command("tmbuild", "-p", "release-pdbs-package.json"))
		CompleteStep(STEP_BUILD_WINDOWS_PACKAGE)
	}

	if !HasCompletedStep(STEP_TEST_WINDOWS_PACKAGE) {
		Run(exec.Command("build/the-machinery/bin/simple-3d.exe"))
		Run(exec.Command("build/the-machinery/bin/simple-draw.exe"))
		Run(exec.Command("build/the-machinery/bin/the-machinery.exe"))
		CompleteStep(STEP_TEST_WINDOWS_PACKAGE)
	}
}

func minorRelease() {
	release := ReadSetting("Minor release number (M.m.p)")
	major := release[0 : len(release)-2]

	if !HasCompletedStep(STEP_CHECK_OUT_SOURCE) {
		Run(exec.Command("git", "checkout", "release/"+major))
		os.Chdir("../sample-projects")
		Run(exec.Command("git", "checkout", "release-"+major))
		os.Chdir("../themachinery")
		CompleteStep(STEP_CHECK_OUT_SOURCE)
	}

	if !HasCompletedStep(STEP_UPDATE_VERSION_NUMBERS) {
		Instruction("Update version numbers in the_machinery.h and *-package.json.")
		CompleteStep(STEP_UPDATE_VERSION_NUMBERS)
	}

	buildWindowsPackage()

	if !HasCompletedStep(STEP_COMMIT_CHANGES) {
		Run(exec.Command("git", "commit", "-a", "-m", "Release "+release))
		Run(exec.Command("git", "push"))
		CompleteStep(STEP_COMMIT_CHANGES)
	}
}

func main() {
	minorRelease()
}
