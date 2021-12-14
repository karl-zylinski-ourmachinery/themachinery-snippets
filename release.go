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

func HotFixLink(version string) string {
	return strings.ReplaceAll(version, ".", "")
}

func buildWindowsPackage() {
	const STEP_CLEAN = "Clean directory"
	if !HasCompletedStep(STEP_CLEAN) {
		Run(exec.Command("tmbuild", "--clean"))
		CompleteStep(STEP_CLEAN)
	}

	const STEP_BUILD_WINDOWS_PACKAGE = "Build Windows package"
	if !HasCompletedStep(STEP_BUILD_WINDOWS_PACKAGE) {
		Run(exec.Command("tmbuild", "-p", "release-package.json"))
		Run(exec.Command("tmbuild", "-p", "release-pdbs-package.json"))
		CompleteStep(STEP_BUILD_WINDOWS_PACKAGE)
	}

	const STEP_TEST_WINDOWS_PACKAGE = "Test Windows package"
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

	const STEP_UPLOAD_WINDOWS_TO_DROPBOX = "Upload Windows package to Dropbox"
	if !HasCompletedStep(STEP_UPLOAD_WINDOWS_TO_DROPBOX) {
		// TODO: Get Dropbox path from user settings somehow...
		dropbox := "C:/Users/nikla/Dropbox (Our Machinery)/Our Machinery Everybody"
		dir := path.Join(dropbox, "releases/early-adopter", Major(version))
		CopyFileToDir(windowsPackage, dir)
		CopyFileToDir(windowsPdbsPackage, dir)
		CompleteStep(STEP_UPLOAD_WINDOWS_TO_DROPBOX)
	}

	const STEP_UPLOAD_WINDOWS_TO_WEBSITE = "Upload Windows package to website"
	if !HasCompletedStep(STEP_UPLOAD_WINDOWS_TO_WEBSITE) {
		password := ReadSetting("Website password")
		dir := "public_html/releases/" + Major(version)
		UploadFileToWebsiteDir(windowsPackage, dir, password)
		CompleteStep(STEP_UPLOAD_WINDOWS_TO_WEBSITE)
	}
}

func hotfixRelease() {
	version := ReadSetting("Hotfix version number (M.m.p)")

	const STEP_CHECK_OUT_SOURCE = "Check out source code"
	if !HasCompletedStep(STEP_CHECK_OUT_SOURCE) {
		Run(exec.Command("git", "checkout", "release/"+Major(version)))
		os.Chdir("../sample-projects")
		Run(exec.Command("git", "checkout", "release-"+Major(version)))
		os.Chdir("../themachinery")
		CompleteStep(STEP_CHECK_OUT_SOURCE)
	}

	// TODO: Automate this step.
	const STEP_UPDATE_VERSION_NUMBERS = "Update version numbers"
	ManualStep(STEP_UPDATE_VERSION_NUMBERS, "Update version numbers in the_machinery.h and *-package.json.")

	buildWindowsPackage()
	uploadWindowsPackage(version)

	const STEP_COMMIT_CHANGES = "Commit changes"
	if !HasCompletedStep(STEP_COMMIT_CHANGES) {
		Run(exec.Command("git", "commit", "-a", "-m", "Release "+version))
		Run(exec.Command("git", "push"))
		CompleteStep(STEP_COMMIT_CHANGES)
	}

	ManualStep("Build on Linux", "Reboot to Linux and run the build script there.")

	// TODO: Automate this step.
	ManualStep("Update website links", "Update the links on the download page with the links to the new project.")

	// TODO: Automate this step.
	ManualStep("Add Release Notes", "Add Release Notes for the hot fix release.")

	const STEP_VERIFY_WEBSITE = "Verify website"
	if !HasCompletedStep(STEP_VERIFY_WEBSITE) {
		os.Chdir("../ourmachinery.com")
		hugoServe := exec.Command("hugo-80", "serve")
		hugoServe.Stdout = os.Stdout
		hugoServe.Stderr = os.Stderr
		err := hugoServe.Start()
		if err != nil {
			log.Fatal(err)
		}
		Run(exec.Command("rundll32", "url.dll,FileProtocolHandler", "http://localhost:1313/"))
		ManualStep(STEP_VERIFY_WEBSITE, "Verify that website is working")
		hugoServe.Process.Kill()
		os.Chdir("../themachinery")
		CompleteStep(STEP_VERIFY_WEBSITE)
	}

	const BUILD_WEBSITE = "Build website"
	if !HasCompletedStep(BUILD_WEBSITE) {
		os.Chdir("../ourmachinery.com")
		Run(exec.Command("hugo-80"))
		os.Chdir("../themachinery")
		CompleteStep(BUILD_WEBSITE)
	}

	const COMMIT_WEBSITE = "Commit website"
	if !HasCompletedStep(COMMIT_WEBSITE) {
		os.Chdir("../ourmachinery.com")
		exec.Command("git", "gui").Run()
		os.Chdir("../themachinery")
		ManualStep(COMMIT_WEBSITE, "Review and commit website changes")
	}

	const UPLOAD_WEBSITE = "Upload website"
	if !HasCompletedStep(UPLOAD_WEBSITE) {
		password := ReadSetting("Website password")
		os.Chdir("../ourmachinery.com/bin")
		Run(exec.Command("go", "run", "upload.go", "-password", password))
		Run(exec.Command("git", "push"))
		os.Chdir("../../themachinery")
		CompleteStep(UPLOAD_WEBSITE)
	}

	const PUSH_TAGS = "Push tags"
	if !HasCompletedStep(PUSH_TAGS) {
		Run(exec.Command("git", "tag", "release-"+version))
		Run(exec.Command("git", "push", "--tags"))
		CompleteStep(PUSH_TAGS)
	}

	const MERGE_TO_MASTER = "Merge to master"
	if !HasCompletedStep(MERGE_TO_MASTER) {
		os.Chdir("../sample-projects")
		Run(exec.Command("git", "checkout", "master"))
		os.Chdir("../themachinery")

		Run(exec.Command("git", "checkout", "master"))
		Run(exec.Command("git", "merge", "release/"+Major(version)))
		Run(exec.Command("git", "push"))
		CompleteStep(MERGE_TO_MASTER)
	}

	const UPDATE_DOWNLOADS_CONFIGS = "Update themachinery/the-machinery-downloads-configs.json"
	if !HasCompletedStep(UPDATE_DOWNLOADS_CONFIGS) {
		dropbox := "C:/Users/nikla/Dropbox (Our Machinery)/Our Machinery Everybody"
		dir := path.Join(dropbox, "releases/early-adopter", Major(version))
		windowsPackage := path.Join(dir, "the-machinery-"+version+"-windows.zip")
		linuxPackage := path.Join(dir, "the-machinery-"+version+"-linux.zip")
		windowsStat, err := os.Stat(windowsPackage)
		if err != nil {
			log.Fatal(err)
		}
		linuxStat, err := os.Stat(linuxPackage)
		if err != nil {
			log.Fatal(err)
		}

		s := `
        {
            "platform": "windows",
            "version": "VERSION",
            "download": "https://ourmachinery.com/releases/MAJOR/the-machinery-VERSION-windows.zip",
            "releaseNotes": "https://ourmachinery.com/post/release-2021-11#HOTFIXLINK",
            "size": "WINDOWS-SIZE"
        },
        {
            "platform": "linux",
            "version": "VERSION",
            "download": "https://ourmachinery.com/releases/MAJOR/the-machinery-VERSION-linux.zip",
            "releaseNotes": "https://ourmachinery.com/post/release-2021-11#HOTFIXLINK",
            "size": "LINUX-SIZE"
        },`
		s = strings.ReplaceAll(s, "MAJOR", Major(version))
		s = strings.ReplaceAll(s, "VERSION", version)
		s = strings.ReplaceAll(s, "HOTFIXLINK", HotFixLink(version))
		s = strings.ReplaceAll(s, "WINDOWS-SIZE", fmt.Sprintf("%v", windowsStat.Size()))
		s = strings.ReplaceAll(s, "LINUX-SIZE", fmt.Sprintf("%v", linuxStat.Size()))
		fmt.Println(s)
		fmt.Println()
		fmt.Println("Press <Enter> to continue when done...")
		fmt.Scanln()
		CompleteStep(UPDATE_DOWNLOADS_CONFIGS)
	}

	const UPLOAD_DOWNLOADS_CONFIGS = "Upload downloads configs"
	if !HasCompletedStep(UPLOAD_DOWNLOADS_CONFIGS) {
		Run(exec.Command("tmbuild"))
		password := ReadSetting("Website password")
		dir := "public_html"
		UploadFileToWebsiteDir("the_machinery/the-machinery-downloads-config.json", dir, password)
		Run(exec.Command("bin/Debug/the-machinery.exe"))
		CompleteStep(UPLOAD_DOWNLOADS_CONFIGS)
	}
}

func linuxBuildFromScratch() {
	version := ReadSetting("Hotfix version number (M.m.p)")
	user := ReadSetting("GitHub user")
	token := ReadSetting("GitHub Access Token")

	os.Chdir("..")
	os.Mkdir("themachinery", 0755)
	os.Chdir("themachinery")
	os.Setenv("TM_OURMACHINERY_COM_DIR", "../ourmachinery.com")
	os.Setenv("TM_SAMPLE_PROJECTS_DIR", "../sample-projects")

	const STEP_CLONE_REPOSITORY = "Clone repository"
	if !HasCompletedStep(STEP_CLONE_REPOSITORY) {
		// Clone main repository
		Run(exec.Command("git", "clone", "https://"+user+":"+token+"@github.com/OurMachinery/themachinery.git", "."))
		Run(exec.Command("git", "checkout", "release/"+Major(version)))

		// Fake ourmachinery.com dir
		os.Mkdir("../ourmachinery.com", 0755)

		// Sample projects
		os.Chdir("..")
		Run(exec.Command("git", "clone", "https://"+user+":"+token+"@github.com/OurMachinery/sample-projects.git"))
		os.Chdir("sample-projects")
		Run(exec.Command("git", "checkout", "release-"+Major(version)))
		os.Chdir("../themachinery")

		CompleteStep(STEP_CLONE_REPOSITORY)
	}

	const STEP_INSTALL_BUILD_LIBRARIES = "Install build libraries"
	if !HasCompletedStep(STEP_INSTALL_BUILD_LIBRARIES) {
		Run(exec.Command("/bin/sh", "-c", "sudo sed -i '1 ! s/restricted/restricted universe multiverse/g' /etc/apt/sources.list"))
		Run(exec.Command("/bin/sh", "-c", "sudo apt update"))
		Run(exec.Command("/bin/sh", "-c", "sudo apt -y install git make clang libasound2-dev libxcb-randr0-dev libxcb-util0-dev libxcb-ewmh-dev"))
		Run(exec.Command("/bin/sh", "-c", "sudo apt -y install libxcb-icccm4-dev libxcb-keysyms1-dev libxcb-cursor-dev libxcb-xkb-dev libxkbcommon-dev"))
		Run(exec.Command("/bin/sh", "-c", "sudo apt -y install libxkbcommon-x11-dev libtinfo5 libxcb-xrm-dev"))
		CompleteStep(STEP_INSTALL_BUILD_LIBRARIES)
	}

	const STEP_INSTALL_TMBUILD = "Install tmbuild"
	if !HasCompletedStep(STEP_INSTALL_TMBUILD) {
		Run(exec.Command("wget", "-O", "tmbuild", "https://www.dropbox.com/s/h4a0subvm5hzwgf/tmbuild?dl=1"))
		Run(exec.Command("chmod", "u+x", "tmbuild"))
		CompleteStep(STEP_INSTALL_TMBUILD)
	}

	const STEP_BOOTSTRAP_TMBUILD_WITH_LATEST = "Bootstrap tmbuild with latest"
	if !HasCompletedStep(STEP_BOOTSTRAP_TMBUILD_WITH_LATEST) {
		Run(exec.Command("./tmbuild", "--project", "tmbuild", "--no-unit-test"))
		Run(exec.Command("cp", "bin/Debug/tmbuild", "."))
		CompleteStep(STEP_BOOTSTRAP_TMBUILD_WITH_LATEST)
	}

	const STEP_BUILD_LINUX_PACKAGE = "Build Linux package"
	if !HasCompletedStep(STEP_BUILD_LINUX_PACKAGE) {
		Run(exec.Command("./tmbuild", "-p", "release-package.json"))
		Run(exec.Command("./tmbuild", "-p", "release-debug-symbols-package.json"))
		CompleteStep(STEP_BUILD_LINUX_PACKAGE)
	}

	const STEP_TEST_LINUX_PACKAGE = "Test Linux package"
	if !HasCompletedStep(STEP_TEST_LINUX_PACKAGE) {
		Run(exec.Command("build/the-machinery/bin/simple-3d"))
		Run(exec.Command("build/the-machinery/bin/simple-draw"))
		Run(exec.Command("build/the-machinery/bin/the-machinery"))
		CompleteStep(STEP_TEST_LINUX_PACKAGE)
	}

	linuxPackage := "build/the-machinery-" + version + "-linux.zip"
	linuxSymbolsPackage := "build/the-machinery-debug-symbols-" + version + "-linux.zip"

	const STEP_UPLOAD_LINUX_TO_DROPBOX = "Upload Linux package to Dropbox"
	if !HasCompletedStep(STEP_UPLOAD_LINUX_TO_DROPBOX) {
		Run(exec.Command("/bin/sh", "-c", "firefox https://www.dropbox.com/work/Our%20Machinery%20Everybody/releases/early-adopter/2021.11"))
		ManualStep(STEP_UPLOAD_LINUX_TO_DROPBOX, "Upload "+linuxPackage+" and "+linuxSymbolsPackage+"to Dropbox")
	}

	const STEP_UPLOAD_LINUX_TO_WEBSITE = "Upload Linux to website"
	if !HasCompletedStep(STEP_UPLOAD_LINUX_TO_WEBSITE) {
		password := ReadSetting("Website password")
		dir := "public_html/releases/" + Major(version)
		UploadFileToWebsiteDir(linuxPackage, dir, password)
		CompleteStep(STEP_UPLOAD_LINUX_TO_WEBSITE)
	}
}

func main() {
	os.Chdir("..")
	hotfixRelease()
	// linuxBuildFromScratch()
}
