#!/usr/bin/env bash

#set -e
#set -x

readonly FIND="${FIND:-/usr/bin/find}"
readonly GOFMT="${GOFMT:-/usr/bin/gofmt}"
readonly GREP="${GREP:-/bin/grep}"

# If we're not running on CircleCI, then allow a directory to be passed in via
# the command line
if [[ -z "$CIRCLE_WORKING_DIRECTORY" ]]; then
	CIRCLE_WORKING_DIRECTORY="$1"
	if [[ -z "$CIRCLE_WORKING_DIRECTORY" ]]; then
		echo "CIRCLE_WORKING_DIRECTORY must be defined. Bailing." >&2
		exit 1
	fi
fi

cd $CIRCLE_WORKING_DIRECTORY

different_files="$($GOFMT -l . | $GREP -v ^vendor)"

for file in $different_files; do
	echo
	$GOFMT -d $file
done

# Fail if gofmt identifies diff files
if [[ -n "$different_files" ]]; then
	exit 1
else
	exit 0
fi

