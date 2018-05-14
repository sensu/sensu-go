#!/usr/bin/env bash

function die() {
  echo $*
  exit 1
}

# Initialize coverage.out
echo "mode: count" > coverage.out

# Initialize error tracking
ERROR=""

declare -a packages=('' \
    'cmd/gostatsd' 'cmd/tester'
    'pkg/backends' 'pkg/backends/sender'
    'pkg/backends/datadog' 'pkg/backends/graphite' 'pkg/backends/null' \
    'pkg/backends/statsdaemon' 'pkg/backends/stdout' \
    'pkg/cloudproviders' 'pkg/cloudproviders/aws' \
    'pkg/fakesocket' 'pkg/statsd' \
    'pkg/statser' 'pkg/cluster/nodes' );

# Test each package and append coverage profile info to coverage.out
for pkg in "${packages[@]}"
do
    go test -v -covermode=count -coverprofile=coverage_tmp.out "github.com/atlassian/gostatsd/$pkg" || ERROR="Error testing $pkg"
    tail -n +2 coverage_tmp.out >> coverage.out 2> /dev/null ||:
done

rm -f coverage_tmp.out

if [ ! -z "$ERROR" ]
then
    die "Encountered error, last error was: $ERROR"
fi
