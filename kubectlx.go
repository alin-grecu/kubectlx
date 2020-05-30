package main

import (
	"bufio"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

var KUBECTL_DEFAULT_PATH string = "/usr/local/bin/kubectl"
var KUBECTL_DOWNLOAD_URL_BASE string = "https://storage.googleapis.com/kubernetes-release/release/v"
var KUBECTL_DEFAULT_VERSION string = "0.0.0"

type Kubectl struct {
	Path    string
	Version string
}

func (k *Kubectl) Exists() bool {
	log.Println("Entering Exists()")
	_, err := os.Stat(k.Path)

	if err != nil {
		return false
	}
	return true
}

func (k *Kubectl) Switch(destination string) (bool, error) {
	log.Println("Entering Switch()")
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

func (k *Kubectl) GetVersion() string {
	log.Println("Entering GetVersion()")
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

func (k *Kubectl) AskConfirmation(question string) (bool, error) {
	log.Println("Entering AskConfirmation()")
	reader := bufio.NewReader(os.Stdin)

	for {
		log.Printf("%s [y/n]:", question)

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

func (k *Kubectl) Download() (bool, error) {
	log.Println("Entering Download()")
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

func Parse(argv []string) (string, error) {
	log.Println("Entering Parse()")
	if len(argv) == 0 {
		return "", errors.New("Too few arguments")
	} else if len(argv) > 1 {
		return "", errors.New("Too many arguments")
	}
	return argv[0], nil
}

func Check(isTrue bool, err error) bool {
	log.Println("Entering Check()")
	if err != nil {
		log.Fatal(err)
		return false
	}
	if isTrue {
		return true
	} else {
		return false
	}
}

func main() {
	// Parse arguments
	version, err := Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

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
		log.Println("Default exists. Saving default.")
		if Check(current.Switch(current.Path + "-" + current.Version)) {
			log.Println("Saved default.")
		}
	}
	if desired.Exists() {
		if current.Version != desired.Version {
			desired.Switch(KUBECTL_DEFAULT_PATH)
			log.Println("Saved version as default.")
		} else {
			log.Println("Identical versions")
		}
	} else {
		if Check(desired.AskConfirmation("Do you want to download this version?")) {
			if Check(desired.Download()) {
				if desired.Version != desired.GetVersion() {
					err := os.Remove(desired.Path)
					log.Fatalln("Invalid version")
					if err != nil {
						log.Fatal(err)
					}
				} else {
					desired.Switch(KUBECTL_DEFAULT_PATH)
					log.Println("Saved version as default.")
				}
			}
		}
	}
}
