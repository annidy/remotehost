package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/akamensky/argparse"
)

func rhost() string {
	file, err := ioutil.ReadFile("/etc/hosts")
	check(err)
	return string(file)
}

func whost(host string) error {
	error := ioutil.WriteFile("/etc/hosts", []byte(host), os.ModePerm)
	return error
}

func delhost(host, name string) string {
	lines := strings.Split(host, "\n")
	result := make([]string, 0)
	found := false
	for _, line := range lines {
		if line == "# "+name+" Start" {
			found = true
		}
		if !found {
			result = append(result, line)
		}
		if line == "# "+name+" End" {
			found = false
		}
	}
	return strings.Join(result, "\n")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	parser := argparse.NewParser("romotehost", "Update host from url")

	dry := parser.Flag("d", "dry", &argparse.Options{Required: false, Help: "print only"})

	url := parser.String("u", "url", &argparse.Options{Required: true, Help: "url"})

	name := parser.String("n", "name", &argparse.Options{Required: true, Help: "rule name"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	resp, err := http.Get(*url)
	if err != nil {
		panic(err)

	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)

	rule := strings.Trim(*name, " ")
	hostTxt := rhost()
	hostTxt = delhost(hostTxt, rule)
	hostTxt = hostTxt + "\n" +
		"# " + rule + " Start\n" +
		string(content) + "\n" +
		"# " + rule + " End\n"

	if *dry {
		fmt.Printf(hostTxt)
		os.Exit(0)
	}
	err = whost(hostTxt)
	check(err)
}