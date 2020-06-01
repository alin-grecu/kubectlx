package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

var BIN_BASE = "/usr/local/bin"
var KUBECTL_DEFAULT_PATH string = "/usr/local/bin/kubectl"
var KUBECTL_DOWNLOAD_URL_BASE string = "https://storage.googleapis.com/kubernetes-release/release/v"

type Kubectl struct {
	Path    string
	Version string
}

// Checks if k.Path exists
func (k *Kubectl) Exists() bool {
	_, err := os.Stat(k.Path)

	if err != nil {
		return false
	}
	return true
}

// Moves file from source to destination, essentialy a copy-paste function
func (k *Kubectl) Switch(destination string) (bool, error) {
	from, err := os.Open(k.Path)

	if err != nil {
		return false, err
	}
	defer from.Close()

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0755)

	if err != nil {
		return false, err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		return false, err
	}
	return true, nil
}

// Returns version number {major}-{minor}-{patch} as string
func (k *Kubectl) GetVersion() string {
	// Execute kubectl to get version info
	cmd := exec.Command(k.Path, "version", "--client=true", "--short")
	out, _ := cmd.CombinedOutput()

	// Parse output to extract version
	parser := regexp.MustCompile(`(Client Version: v)(.*)`)
	result := parser.FindStringSubmatch(string(out))

	// Parsing failed if less than 3 elements are returned
	if len(result) < 3 {
		return ""
	}

	// Return version as string e.g 1.12.0
	return result[2]
}

// Just a simple prompt for user confirmation before Download takes place
func (k *Kubectl) AskConfirmation(question string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	for {
		// This is a clean workaround to using log.Printf which adds a new line
		fmt.Printf("%s %s [y/n]: ", time.Now().Format("2006/01/02 15:04:05"), question)
		response, err := reader.ReadString('\n')
		if err != nil {
			return false, err
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true, nil
		} else if response == "n" || response == "no" {
			return false, nil
		}
	}
}

// Handles kubectl file download and permissions
func (k *Kubectl) Download() (bool, error) {
	url := KUBECTL_DOWNLOAD_URL_BASE + k.Version + "/bin/" + runtime.GOOS + "/" + runtime.GOARCH + "/" + "kubectl"

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(k.Path)
	if err != nil {
		return false, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return false, err
	}

	// Set permissions
	err = os.Chmod(k.Path, 0755)
	if err != nil {
		return false, err
	}

	return true, nil
}

func FindFiles(filesPath string) []string {
	matches, err := filepath.Glob(filesPath)
	if err != nil {
		log.Fatalln(err)
	}

	if len(matches) == 0 {
		return nil
	}
	return matches
}

func ListVersions() {
	var versions []string
	versionPaths := FindFiles(KUBECTL_DEFAULT_PATH + "-*")
	if versionPaths == nil {
		log.Fatalln("No versions installed with kubectlx")
	}
	for _, v := range versionPaths {
		versions = append(versions, strings.Trim(v, "/usr/local/bin/kubectl-"))
	}
	for _, v := range versions {
		fmt.Println(v)
	}
}

func Parse(argv []string) (string, error) {
	if len(argv) == 0 {
		return "", errors.New("Too few arguments")
	} else if len(argv) > 1 {
		return "", errors.New("Too many arguments")
	}
	return argv[0], nil
}

// This is a handler for some functions returning both bool and err
// it makes main() a bit more readable
func Check(isTrue bool, err error) bool {
	if err != nil {
		log.Fatal(err)
	}
	return isTrue
}

func main() {
	// Parse arguments
	option, err := Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if option == "list" {
		ListVersions()
		os.Exit(0)
	}

	version := option

	// Initialize
	var current, desired Kubectl

	// Set current
	current.Path = KUBECTL_DEFAULT_PATH
	current.Version = current.GetVersion()

	// Set desired
	desired.Version = version
	desired.Path = KUBECTL_DEFAULT_PATH + "-" + desired.Version

	// Check if kubectl default exists
	if current.Exists() {
		Check(current.Switch(current.Path + "-" + current.Version))
	}
	if !desired.Exists() {
		if Check(desired.AskConfirmation("You do not have this version. Do you want to download it?")) {
			if Check(desired.Download()) {
				if desired.Version != desired.GetVersion() {
					err := os.Remove(desired.Path)
					log.Fatalf("Version %s is invalid. I have removed %s", desired.Version, desired.Path)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		} else {
			os.Exit(0)
		}
	}
	desired.Switch(KUBECTL_DEFAULT_PATH)
}
