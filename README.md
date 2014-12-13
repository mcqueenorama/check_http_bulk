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



