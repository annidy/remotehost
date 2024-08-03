package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
	"log"

	"github.com/akamensky/argparse"
)

func readfile() string {
	file, err := ioutil.ReadFile("/etc/hosts")
	check(err)
	return string(file)
}

func writefile(host string) error {
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
	var i = len(result)-1
	for ; i > 0; i-- {
		if len(result[i-1]) != 0 {
			break
		}
	}
	result = result[:i]
	return strings.Join(result, "\n")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func flush_dns() {
	cmd := exec.Command("dscacheutil", "-flushcache")
	err := cmd.Run()
    if err != nil {
		log.Fatal("dscacheutil", err)
	}
	cmd = exec.Command("killall", "-HUP", "mDNSResponder")
	err = cmd.Run()
	if err != nil {
		log.Fatal("kill mDNSResponder", err)
	}
	log.Println("flush dns")
}

func main() {

	parser := argparse.NewParser("romotehost", "Update host from url")
	dry := parser.Flag("d", "dry", &argparse.Options{Required: false, Help: "print only"})
	url := parser.String("u", "url", &argparse.Options{Required: false, Help: "fetch url"})
	name := parser.String("n", "name", &argparse.Options{Required: true, Help: "rule name"})
	rm := parser.Flag("r", "rm", &argparse.Options{Required: false, Help: "remove rule"})
	itv := parser.Int("i", "interval", &argparse.Options{Required: false, Help: "minutes of next fetch"})
	verbose := parser.Flag("v", "verbose", &argparse.Options{Required: false, Help: "verbose"})

	// Parse input
	err := parser.Parse(os.Args)
	if err != nil {
		// In case of error print error and print usage
		// This can also be done by passing -h or --help flags
		fmt.Print(parser.Usage(err))
	}

	rule := strings.Trim(*name, " ")

	// 监听SIGNUP命令
	signCh := make(chan os.Signal, 1)
	signal.Notify(signCh, os.Interrupt, syscall.SIGTERM)
	wg := sync.WaitGroup{}
	wg.Add(1)
	worker := func() {
		defer wg.Done()
		for {
			hostTxt := readfile()
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

				if resp.StatusCode != 200 {
					log.Fatal("Server bad")
				}

				hostTxt = delhost(hostTxt, rule)
				hostTxt = hostTxt + "\n\n" +
					"# " + rule + " Start\n" +
					string(content) + "\n" +
					"# " + rule + " End\n"
			} else if *rm {
				hostTxt = delhost(hostTxt, rule)
			}

			if *dry {
				fmt.Println(hostTxt)
			} else {
				err = writefile(hostTxt)
				check(err)
				fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "update host successfuly")
				flush_dns()
			}

			if *itv != 0 {
				select {
				case <-signCh:
					fmt.Println("Interupt by user")
					if *rm {
						*url = ""
						*itv = 0
					} else {
						break
					}
				case <-time.After(time.Duration(*itv*60) * time.Second):
					continue
				}
			} else {
				break
			}
		}
	}
	go worker()
	wg.Wait()
}
