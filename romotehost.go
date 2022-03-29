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
	result := make([]string, 0, len(lines))
	found := false
	for _, line := range lines {
		if line == "# "+name+" Start" {
			found = true
		}
		if line == "# "+name+" End" {
			found = false
		} else if !found {
			result = append(result, line)
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

	url := parser.String("u", "url", &argparse.Options{Required: false, Help: "url"})

	name := parser.String("n", "name", &argparse.Options{Required: true, Help: "rule name"})

	rm := parser.Flag("r", "rm", &argparse.Options{Required: false, Help: "remove"})

	verbose := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "verbose"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	rule := strings.Trim(*name, " ")
	hostTxt := rhost()

	if *url != "" {
		resp, err := http.Get(*url)
		if err != nil {
			panic(err)
		}
		defer resp.Body.Close()
		content, _ := ioutil.ReadAll(resp.Body)

		if *verbose {
			fmt.Println(string(content))
		}

		hostTxt = delhost(hostTxt, rule)
		hostTxt = hostTxt + "\n" +
			"# " + rule + " Start\n" +
			string(content) + "\n" +
			"# " + rule + " End\n"
	}

	if *rm {
		hostTxt = delhost(hostTxt, rule)
	}

	if *dry {
		fmt.Println(hostTxt)
		os.Exit(0)
	}
	err = whost(hostTxt)
	check(err)
}
