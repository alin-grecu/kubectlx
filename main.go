package main

import (
	"bufio"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type Kubectl struct {
	Path    string
	Version string
}

func (k *Kubectl) GetVersion() string {
	// Call kubectl to get version
	cmd := exec.Command(k.Path, "version", "--client=true", "--short")
	out, _ := cmd.CombinedOutput()

	// Parse output to extract major, minor and patch
	parser := regexp.MustCompile(`(Client Version: v)(.*)`)
	result := parser.FindStringSubmatch(string(out))

	if len(result) == 0 {
		return ""
	}

	return result[2]
}

func (k *Kubectl) CheckBinExists() bool {
	_, err := os.Stat(k.Path)
	if os.IsNotExist(err) {
		log.Println(err)
		return false
	}
	return true
}

func (k *Kubectl) SetPath(save bool) {
	source := k.Path
	destination := "/usr/local/bin/kubectl"

	if save {
		source = "/usr/local/bin/kubectl"
		destination = k.Path
	}

	from, err := os.Open(source)
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	to, err := os.OpenFile(destination, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	if err != nil {
		log.Fatal(err)
	}
}

func (k *Kubectl) Download() error {
	log.Printf("Downloading kubectl version %s\n", k.Version)
	url := "https://storage.googleapis.com/kubernetes-release/release/v" +
		k.Version +
		"/bin/" +
		runtime.GOOS +
		"/amd64/kubectl"

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(k.Path)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)

	// Set permissions
	err = os.Chmod(k.Path, 0755)

	return err
}

func (k *Kubectl) ValidateDownload() {
	// Check version
	downloaded_version := k.GetVersion()
	if k.Version != downloaded_version {
		err := os.Remove(k.Path)
		if err != nil {
			log.Fatal(err)
		}
		log.Fatalf("%s is invalid or the downloaded version doesn't exist", k.Path)
	}
}

func parseArgs(argv []string) string {
	if len(argv) == 0 {
		log.Println("Too few arguments")
		os.Exit(1)
	}
	version := argv[0]
	return version
}

func askConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		log.Printf("%s [y/n]:", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func main() {

	// Parse arguments
	version := parseArgs(os.Args[1:])

	// Initialize current and desired structs
	var current, desired Kubectl
	current.Path = "/usr/local/bin/kubectl"

	// Set desired version
	desired.Version = version

	// Check if kubectl exists
	isPresent := current.CheckBinExists()
	if isPresent {

		// Save current kubectl
		current.Version = current.GetVersion()
		current.SetPath(true)

		// Check if desired version is the same as current
		if desired.Version == current.Version {
			log.Fatalf("You are already using version %s", desired.Version)
		}
	}

	// Set desired version path
	desired.Path = "/usr/local/bin/kubectl" + "-" + desired.Version

	// Check if desired version path exists
	isPresent = desired.CheckBinExists()
	if !isPresent {

		// ask confirmation to download missing version
		response := askConfirmation("Do you want to download this version?")
		if response {
			err := desired.Download()
			if err != nil {
				log.Println(err)
			} else {
				desired.ValidateDownload()
			}
		} else {
			os.Exit(1)
		}
	}

	// Set version
	desired.SetPath(false)
}
