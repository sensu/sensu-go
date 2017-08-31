#!/bin/sh

# safe_rpm_arch takes a go-compatible arch and will return an rpm-compatible arch
# e.g. amd64 -> x86_64
#
safe_rpm_arch() {
    if [ $1 = "amd64" ]; then
	echo "x86_64"
    fi
}

# safe_rpm_version will return a version string that is rpm-compatible
# e.g.
#
safe_rpm_version() {
    echo "Not implemented yet"
    exit 1
}
