package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/jlaffaye/ftp"
)

const settingsFile = "releaseBuild.json"

const STEP_CHECK_OUT_SOURCE = "Check out source code"
const STEP_CLEAN = "Clean directory"
const STEP_BUILD_WINDOWS_PACKAGE = "Build Windows package"
const STEP_TEST_WINDOWS_PACKAGE = "Test Windows package"
const STEP_UPDATE_VERSION_NUMBERS = "Update version numbers"
const STEP_COMMIT_CHANGES = "Commit changes"
const STEP_UPLOAD_WINDOWS_TO_DROPBOX = "Upload Windows package to Dropbox"
const STEP_UPLOAD_WINDOWS_TO_WEBSITE = "Upload Windows package to website"

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

func ManualStep(s, details string) {
	if !HasCompletedStep(s) {
		fmt.Println(details)
		fmt.Println()
		fmt.Println("Press <Enter> to continue when done...")
		fmt.Scanln()
		CompleteStep(s)
	}
}

func CopyFileToDir(srcFile, dir string) {
	dstFile := path.Join(dir, path.Base(srcFile))
	src, err := os.Open(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()
	dst, err := os.Create(dstFile)
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()
	_, err = io.Copy(dst, src)
	if err != nil {
		log.Fatal(err)
	}
}

func UploadFileToWebsiteDir(srcFile, dir, password string) {
	c, err := ftp.Dial("160.153.16.15:21")
	if err != nil {
		log.Fatal(err)
	}

	err = c.Login("ourmachinery", password)
	if err != nil {
		log.Fatal(err)
	}

	err = c.ChangeDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open(srcFile)
	if err != nil {
		log.Fatal(err)
	}
	base := path.Base(srcFile)
	err = c.Stor(base, f)

	c.Quit()
}

func Major(version string) string {
	fields := strings.Split(version, ".")
	return fields[0] + "." + fields[1]
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

func uploadWindowsPackage(version string) {
	windowsPackage := "build/the-machinery-" + version + "-windows.zip"
	windowsPdbsPackage := "build/the-machinery-pdbs-" + version + "-windows.zip"

	if !HasCompletedStep(STEP_UPLOAD_WINDOWS_TO_DROPBOX) {
		// TODO: Get Dropbox path from user settings somehow...
		dropbox := "C:/Users/nikla/Dropbox (Our Machinery)/Our Machinery Everybody"
		dir := path.Join(dropbox, "releases/early-adopter", Major(version))
		CopyFileToDir(windowsPackage, dir)
		CopyFileToDir(windowsPdbsPackage, dir)
		CompleteStep(STEP_UPLOAD_WINDOWS_TO_DROPBOX)
	}

	if !HasCompletedStep(STEP_UPLOAD_WINDOWS_TO_WEBSITE) {
		password := ReadSetting("Website password")
		dir := "public_html/releases/" + Major(version)
		UploadFileToWebsiteDir(windowsPackage, dir, password)
		CompleteStep(STEP_UPLOAD_WINDOWS_TO_WEBSITE)
	}
}

func hotfixRelease() {
	version := ReadSetting("Hotfix version number (M.m.p)")

	if !HasCompletedStep(STEP_CHECK_OUT_SOURCE) {
		Run(exec.Command("git", "checkout", "release/"+Major(version)))
		os.Chdir("../sample-projects")
		Run(exec.Command("git", "checkout", "release-"+Major(version)))
		os.Chdir("../themachinery")
		CompleteStep(STEP_CHECK_OUT_SOURCE)
	}

	// TODO: Automate this step.
	ManualStep(STEP_UPDATE_VERSION_NUMBERS, "Update version numbers in the_machinery.h and *-package.json.")

	buildWindowsPackage()
	uploadWindowsPackage(version)

	if !HasCompletedStep(STEP_COMMIT_CHANGES) {
		Run(exec.Command("git", "commit", "-a", "-m", "Release "+version))
		Run(exec.Command("git", "push"))
		CompleteStep(STEP_COMMIT_CHANGES)
	}

	ManualStep("Build on Linux", "Reboot to Linux and run the build script there.")
}

func main() {
	hotfixRelease()
}
