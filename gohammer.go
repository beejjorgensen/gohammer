package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"time"
)

var threadCount uint
var delay time.Duration
var hitCount uint
var url string
var done chan bool
var spread time.Duration

func init() {
	flag.UintVar(&threadCount, "t", 4, "Number of threads to spawn")
	flag.UintVar(&hitCount, "c", 0, "Number of URL hits per thread, zero to unlimit")
	flag.DurationVar(&delay, "d", 100*time.Millisecond, "Delay between web requests")
	flag.DurationVar(&spread, "s", 1*time.Second, "Spread thread launch over time")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] url\n\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

func parseCL() {
	flag.Parse()

	args := flag.Args()

	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	url = args[0]
}

func hitURL() {
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error getting URL: %v", err)
		return
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)
	}
}

func hammer() {
	unlimited := hitCount == 0

	for i := uint(0); i < hitCount || unlimited; i++ {
		hitURL()
		time.Sleep(delay)
	}

	done <- true
}

func main() {
	done = make(chan bool)

	parseCL()

	for i := uint(0); i < threadCount; i++ {
		go hammer()

		if spread > 0 {
			time.Sleep(spread / time.Duration(threadCount))
		}
	}

	// Wait for threads to complete
	for i := uint(0); i < threadCount; i++ {
		<-done
	}
}
