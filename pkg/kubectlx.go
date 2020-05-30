package kubectlx

import (
	"bufio"
	"errors"
	"fmt"
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
var KUBECTL_DOWNLOAD_URL_BASE string = "https://storage.googleapis.com/kubernetes-release/release/"
var KUBECTL_DEFAULT_VERSION string = "0.0.0"

type Kubectl struct {
	Path    string
	Version string
}

func (k *Kubectl) New() {
	if k.Path == "" {
		k.Path = KUBECTL_DEFAULT_PATH
	}
	if k.Version == "" {
		k.Version = KUBECTL_DEFAULT_VERSION
	}
}

func (k *Kubectl) Exists() (bool, error) {
	_, err := os.Stat(k.Path)

	if err != nil {
		return false, err
	}
	return true, nil
}

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

func (k *Kubectl) IsIdentical(desired *Kubectl) (bool, error) {
	if k.GetVersion() != desired.GetVersion() {
		var err error = os.Remove(k.Path)
		return false, err
	}
	return true, nil
}

func (k *Kubectl) AskConfirmation(question string) (bool, error) {
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
	if len(argv) == 0 {
		return "", errors.New("Too few arguments")
	}
	return argv[0], nil
}

func CheckIf(isTrue bool, err error) bool {
	if err != nil {
		log.Fatal(err)
	}
	return true
}

func main() {
	// Parse arguments
	version, err := Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}

	// Initialize
	var current, desired Kubectl
	current.New()
	desired.New()

	// Set desired version
	desired.Version = version

	// Check if kubectl default exists
	if CheckIf(current.Exists()) {
		fmt.Println("File exists")
	}

}
