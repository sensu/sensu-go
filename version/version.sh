#!/bin/sh

##
# This script will output the type of build that this git commit should be
# considered.
#
# * If no git tag exists, it is a nightly build
# * If a git tag exists and the name contains "alpha", it is an alpha build
# * If a git tag exists and the name contains "beta", it is a beta build
# * If a git tag exists and the name contains "rc", it is an rc build
# * If a git tag exists and does not contain the above, it is a stable build
##

script="version.sh"

function build_type() {
    TAG=$(git describe --exact-match --tags HEAD 2> /dev/null || exit 1)

    case "x$TAG" in
        x)
            echo "nightly"
            ;;
        *alpha*)
            echo "alpha"
            ;;
        *beta*)
            echo "beta"
            ;;
        *rc*)
            echo "rc"
            ;;
        *)
            echo "stable"
            ;;
    esac
}

function usage() {
    echo "$script - output various version information"
    echo " "
    echo "$script [options]"
    echo " "
    echo "options:"
    echo "-h, --help                show help"
    echo "-t, --build-type          output the type of build this is"
    exit 0
}

case "$1" in
    -h|--help)
        usage
        ;;
    -t|--build-type)
        build_type
        ;;
    *)
        usage
        ;;
esac
