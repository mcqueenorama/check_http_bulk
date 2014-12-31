package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	// "io"
	"flag"
	"io/ioutil"
	"strings"
	"time"
)

func get(hostname string, port int, path string, auth string, urls bool, verbose bool, timeout int) (rv bool, err error) {

	defer func() {
		if err := recover(); err != nil {
			return
		}
	}()

	rv = true

	if verbose {
		fmt.Fprintf(os.Stderr, "fetching:hostname:%s:\n", hostname)
	}

	res := &http.Response{}

	if urls {

		url := hostname
		res, err = http.Head(url)
		defer res.Body.Close()

		if err != nil {
			fmt.Println(err.Error())
			rv = false
			return
		}

	} else {

		client := &http.Client{Timeout: time.Duration(timeout) * time.Second}

		// had to allocate this or the SetBasicAuth will panic
		headers := make(map[string][]string)
		hostPort := fmt.Sprintf("%s:%d", hostname, port)

		if verbose {

			fmt.Fprintf(os.Stderr, "adding hostPort:%s:%d:path:%s:\n", hostname, port, path)

		}
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

			if verbose {

				fmt.Fprintf(os.Stderr, "Doing auth with:username:%s:password:%s:", up[0], up[1])

			}
			req.SetBasicAuth(up[0], up[1])

		}

		if verbose {

			dump, _ := httputil.DumpRequestOut(req, true)
			fmt.Fprintf(os.Stderr, "%s", dump)

		}

		res, err = client.Do(req)
		if err != nil {
			fmt.Println(err.Error())
			rv = false
			return
		}

		defer res.Body.Close()
		_, err = ioutil.ReadAll(res.Body)

	}

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
	name := "Bulk HTTP Check"
	bad := 0
	total := 0

	// this needs improvement. the number of spaces here has to equal the number of chars in the badHosts append line suffix
	badHosts := []byte("  ")

	verbose := flag.Bool("v", false, "verbose output")
	warn := flag.Int("w", 10, "warning level - number of non-200s or percentage of non-200s (default is numeric not percentage)")
	crit := flag.Int("c", 20, "critical level - number of non-200s or percentage of non-200s (default is numeric not percentage)")
	timeout := flag.Int("t", 2, "timeout in seconds - don't wait.  Do Head requests and don't wait.")
	pct := flag.Bool("pct", false, "interpret warming and critical levels are percentages")
	path := flag.String("path", "", "optional path to append to the input lines including the leading slash - these will not be urlencoded. This is ignored is the urls option is given.")
	file := flag.String("file", "", "input data source: a filename or '-' for STDIN.")
	port := flag.Int("port", 80, "optional port for the http request - ignored if urls is specified")
	urls := flag.Bool("urls", false, "Assume the input data is full urls - its normally a list of hostnames")
	auth := flag.String("auth", "", "Do basic auth with this username:passwd - ignored if urls is specified - make this use .netrc instead")
	checkName := flag.String("name", "", "a name to be included in the check output to distinguish the check output")

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

	The -path is appended to the hostnames to make full URLs for the checks.

	If the -urls option is specified, then the input is assumed a complete URLs, like http://$hostname:$port/$path.

	Examples:

	./someCommand |  ./check_http_bulk  -w 1 -c 2 -path '/api/aliveness-test/%%2F/' -port 15672 -file - -auth zup:nuch 

	./check_http_bulk -urls -file urls.txt

`)

		flag.PrintDefaults()
	}

	flag.Parse()

	if len(flag.Args()) > 0 {

		flag.Usage()
		os.Exit(3)

	}

	// it urls is specified, the input is full urls to be used enmasse and to be url encoded
	if *urls {
		*path = ""
	}

	if *checkName != "" {
		name = *checkName
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(name+" Unknown: ", err)
			os.Exit(3)
		}
	}()

	if file == nil || *file == "" {
		flag.Usage()
		os.Exit(3)
	}

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

		if *verbose {

			fmt.Printf("working on:%s:\n", hostname)

		}

		goodCheck, err := get(hostname, *port, *path, *auth, *urls, *verbose, *timeout)
		if err != nil {

			fmt.Printf("%s get error: %T %s %#v\n", name, err, err, err)
			badHosts = append(badHosts, hostname...)
			badHosts = append(badHosts, ", "...)
			bad++

			continue

		}

		if !goodCheck {
			badHosts = append(badHosts, hostname...)
			badHosts = append(badHosts, ", "...)
			bad++
		}

	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
		status = "Unknown"
		rv = 3
	}

	if *pct {

		ratio := int(float64(bad)/float64(total)*100)

		if *verbose {

			fmt.Fprintf(os.Stderr, "ratio:%d:\n", ratio)

		}

		if ratio >= *crit {
			status = "Critical"
			rv = 2
		} else if ratio >= *warn {
			status = "Warning"
			rv = 1
		}

	} else {

		if bad >= *crit {
			status = "Critical"
			rv = 2
		} else if bad >= *warn {
			status = "Warning"
			rv = 1
		}

	}

	fmt.Printf("%s %s: %d of %d |%s\n", name, status, bad, total, badHosts[:len(badHosts)-2])
	os.Exit(rv)
}
