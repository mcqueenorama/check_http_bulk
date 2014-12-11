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

Next
====

I intend to add a mode where the full url is specified in the input
file, so arbitary batches of urls can be checked in an easy way.

It will be like this with percentages of failures on the alert thresholds:

    cat arbitraryUrls.txt |  ./check_http_bulk  -w 33 -c 66



