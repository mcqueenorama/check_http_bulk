package main

import (
	"bufio"
	"fmt"
	"os"
	"net/http"
	"net/http/httputil"
	"net/url"
	// "io"
	"io/ioutil"
	"flag"
	"time"
	"strings"
)

func get(hostname string, port int, path string, auth string, verbose bool, timeout int) (rv bool, err error) {

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		return
	// 	}
	// }()

	rv = true

	if verbose {
		fmt.Fprintf(os.Stderr, "fetching:hostname:%s:\n", hostname)
	}

	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	// res, err := client.Head(url)

	// req, err := http.NewRequest("HEAD", url, nil)
	
	// if err != nil {
	// 	rv = false
	// 	return
	// }

	// had to allocate this or the SetBasicAuth below causes a panic
    headers := make(map[string][]string)
    hostPort := fmt.Sprintf("%s:%d", hostname, port)
    fmt.Fprintf(os.Stderr, "adding hostPort:%s:%d:path:%s:\n", hostname, port, path)
	req := &http.Request{
		Method: "HEAD",
		// Host:  hostPort,
		URL: &url.URL{
			Host:   hostPort,
			Scheme: "http",
			Opaque: path,
		},
		Header: headers,
	}

    if auth != "" {

    	up := strings.SplitN(auth, ":", 2)
	    fmt.Fprintf(os.Stderr, "Doing auth with:username:%s:password:%s:", up[0], up[1])
		req.SetBasicAuth(up[0], up[1])

    }

    if verbose {

	    dump, _ := httputil.DumpRequestOut(req, true)
	    fmt.Fprintf(os.Stderr, "%s", dump)
    	
    }

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		rv = false
		return
	}

	defer res.Body.Close()
	_, err = ioutil.ReadAll(res.Body)

	// res, err := http.Head(url)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// 	rv = false
	// 	return
	// }

	if verbose {

		fmt.Println(res.Status)
		for k, v := range res.Header {
			fmt.Println(k+":", v)
		}

	}

	if res.StatusCode != http.StatusOK {
		rv = false
	}

	return
}


func main() {

	status := "OK"
	rv := 0
	name := "Bulk HTTP"
	bad := 0
	total := 0

	verbose := flag.Bool("v", false, "verbose output")
	warn := flag.Int("w", 10, "warning level - number of non-200s or percentage of non-200s (default is numeric not percentage)")
	crit := flag.Int("c", 20, "critical level - number of non-200s or percentage of non-200s (default is numeric not percentage)")
	timeout := flag.Int("t", 2, "timeout in seconds - don't wait.  Do Head requests and don't wait.")
	// pct := flag.Bool("pct", false, "interpret warming and critical levels are percentages")
	path := flag.String("path", "", "optional path to append to the stdin lines - these will not be urlencoded. This is ignored is the urls option is given (not implemented yet).")
	file := flag.String("file", "", "optional path to read data from a file instead of stdin.  If its a dash then read from stdin - these will not be urlencoded")
	port := flag.Int("port", 80, "optional port for the http request")
	// bare := flag.Bool("urls", false, "Assume the input data is full urls - its normally a list of hostnames")
	auth := flag.String("auth", "", "Do basic auth with this username:passwd - make this use .netrc instead")

	flag.Usage = func() {

        fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
        fmt.Fprintf(os.Stderr, `
	Read hostnames from a file or STDIN and do a single nagios check over
	them all.  Just check for 200s.  Warning and Critical are either
	percentages of the total, or a regular numeric thresholds.

	The output contains the hostname of any non-200 reporting hosts.

	Skip input lines that are commented out with shell style comments
	like /^#/.

	Do Head requests since we don't care about the content.  Make this
	optional some day.

	Also make this read from a file of urls with a -f option.

	Make the auth configurable from the cli

    	`)

        flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) > 0 {

		flag.Usage()
		os.Exit(3)

	}

	// defer func() {
	// 	if err := recover(); err != nil {
	// 		fmt.Println(name+" Unknown: ", err)
	// 		os.Exit(3)
	// 	}
	// }()

	if file == nil || *file == "" {
		flag.Usage()
		os.Exit(3)
	}

	// inputSource := &io.Reader
	inputSource := os.Stdin

	if (*file)[0] != "-"[0] {

		var err error

		inputSource, err = os.Open(*file)

		if err != nil {

			fmt.Printf("Couldn't open the specified input file:%s:error:%v:\n\n", name, err)
			flag.Usage()
			os.Exit(3)

		}

	}

	scanner := bufio.NewScanner(inputSource)
	for scanner.Scan() {

		total++

		hostname := scanner.Text()

		if hostname[0] == "#"[0] {

			if *verbose {

				fmt.Printf("skipping:%s:\n", hostname)

			}

			continue
		}

		// url := hostname + *path

		if *verbose {

			fmt.Printf("working on:%s:\n", hostname)

		}

		goodCheck, err := get(hostname, *port, *path, *auth, *verbose, *timeout)
		if err != nil {

			fmt.Printf("%s get error: %T %s %#v\n", name, err, err, err)

			// os.Exit(3)
			continue

		}

		if !goodCheck {
			bad++
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}


	if bad >= *crit {
		status = "Critical"
		rv = 1
	} else if bad >= *warn {
		status = "Warning"
		rv = 2
	}

	fmt.Printf("%s %s: %d\n", name, status, bad)
	os.Exit(rv)
}
