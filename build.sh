#!/usr/bin/env bash
set -o pipefail
set -e

# On CircleCI, we want to see what commands are being run, at least for now
test -n "$CIRCLECI" && set -x

eval $(go env)

cmd=${1:-"all"}

RACE=""

set_race_flag() {
    if [ "$GOARCH" == "amd64" ] && [ "$CGO_ENABLED" == "1" ]; then
        RACE="-race"
    fi
}

case "$GOOS" in
    darwin)
        set_race_flag
        ;;
    freebsd)
        set_race_flag
        ;;
    linux)
        set_race_flag
        ;;
    windows)
        set_race_flag
        ;;
esac

unit_test_commands () {
    echo "Running unit tests..."

    go test -v $RACE ./...
    if [ $? -ne 0 ]; then
        echo "Unit testing failed..."
        exit 1
    fi
}

case "$cmd" in
    "none")
        echo "noop"
        ;;
    "unit")
        unit_test_commands
        ;;
    *)
        unit_test_commands
        ;;
esac
