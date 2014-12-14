check_http_bulk
===============

A single nagios checks summarizing a list of urls from file or stdin.

This is useful if you've got a cluster and want to alerting on the
overall health, and the hostnames are coming from some cloud. The
hostnames can be pulled from whatever cloud api and piped into this
for a healthcheck of the cluster.  Something like this:

Check the health of the rabbitmq cluster where the alert thresholds are number of failures:

    ./someCommand |  ./check_http_bulk  -w 1 -c 2 -path '/api/aliveness-test/%2F/' -port 15672 -file - -auth zup:nuch  

or:

    ./check_http_bulk  -w 1 -c 2 -path '/api/aliveness-test/%2F/' -port 15672 -file rabbitHosts.txt -auth zup:nuch

The return is standard Nagios check format, showing failed hosts if there are any:

Hostnames/IPs on STDIN
==================
If working in a cloud provisioned environment its useful to get the ips of a cluster from an api and pipe them into this command to check the health of the cluster as shown above.

Urls on STDIN
=============
For general use its helpful to take a list of interesting urls and pipe them into this check to see how healthy the entire group is:

    ./check_http_bulk -urls -file urls.txt

or:

    cat urls.txt | ./check_http_bulk -urls -file -


Error on Percentage of Failure
==============================

If you want to use percentage of failures instead of an numeric threshold:

    cat arbitraryUrls.txt |  ./check_http_bulk  -pct -w 33 -c 66 -file -

Help
====

	Help:

	bash-3.2$   ./check_http_bulk   -help
Usage of ./check_http_bulk:

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

	./someCommand |  ./check_http_bulk  -w 1 -c 2 -path '/api/aliveness-test/%2F/' -port 15672 -file - -auth zup:nuch 

	./check_http_bulk -urls -file urls.txt

  -auth="": Do basic auth with this username:passwd - ignored if urls is specified - make this use .netrc instead
  -c=20: critical level - number of non-200s or percentage of non-200s (default is numeric not percentage)
  -file="": input data source: a filename or '-' for STDIN.
  -name="": a name to be included in the check output to distinguish the check output
  -path="": optional path to append to the input lines including the leading slash - these will not be urlencoded. This is ignored is the urls option is given.
  -pct=false: interpret warming and critical levels are percentages
  -port=80: optional port for the http request - ignored if urls is specified
  -t=2: timeout in seconds - don't wait.  Do Head requests and don't wait.
  -urls=false: Assume the input data is full urls - its normally a list of hostnames
  -v=false: verbose output
  -w=10: warning level - number of non-200s or percentage of non-200s (default is numeric not percentage)




