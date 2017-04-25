// gohammer
//
// Hammer a webserver repeatedly
//
// Unlicensed by Brian "Beej Jorgensen" Hall <beej@beej.us> 2017
// http://unlicense.org/
//
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

// threadCount is the numeber of threads to run simultaneously
var threadCount uint

// Delay is how long each thread should delay between hits
var delay time.Duration

// hitCount is how many times to hit a URL per thread (0 to unlimit)
var hitCount uint

// url is the URL to hit
var url string

// done is used to report completion to main thread
var done chan bool

// spread is a time period to spread out the launch of the threads
var spread time.Duration

// init
func init() {
	flag.UintVar(&threadCount, "t", 4, "Number of threads to spawn")
	flag.UintVar(&hitCount, "c", 0, "Number of URL hits per thread, zero to unlimit (default 0)")
	flag.DurationVar(&delay, "d", 100*time.Millisecond, "Delay between web requests")
	flag.DurationVar(&spread, "s", 1*time.Second, "Spread thread launch over time")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: %s [options] url\n\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
	}
}

// parseCL parses the command line
func parseCL() {
	flag.Parse()

	args := flag.Args()

	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	url = args[0]
}

// hitURL hits a URL and reads the result
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

// hammer repeatedly hits a URL (used as a goroutine)
func hammer() {
	unlimited := hitCount == 0

	for i := uint(0); i < hitCount || unlimited; i++ {
		hitURL()
		time.Sleep(delay)
	}

	done <- true
}

// main
func main() {
	done = make(chan bool)

	// Parse command line
	parseCL()

	// Launch all threads
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
