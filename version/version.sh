#!/bin/sh

script="version.sh"

# build_type will output the type of build that this git commit should be
# considered.
function build_type() {
    TAG=$(git describe --exact-match --tags HEAD 2> /dev/null || exit 1)

    case "x$TAG" in
        x)
            # no tag exists, this is a nightly build
            echo "nightly"
            ;;
        *alpha*)
            # tag exists and contains alpha, this is an alpha build
            echo "alpha"
            ;;
        *beta*)
            # tag exists and contains beta, this is a beta build
            echo "beta"
            ;;
        *rc*)
            # tag exists and contains rc, this is a release candidate build
            echo "rc"
            ;;
        *)
            # tag exists but does not contain any of the above, this is a stable build
            echo "stable"
            ;;
    esac
}

# iteration will output an iteration number based on what type of build the git
# sha represents and the ci platform it is running on.
function iteration() {
    if [ $(build_type) == "nightly" ]; then
        # get the iteration from environment variables
        if [[ $TRAVIS == "true" ]]; then
            echo $TRAVIS_BUILD_NUMBER
        elif [[ $APPVEYOR == "true" ]]; then
            echo $APPVEYOR_BUILD_NUMBER
        elif [[ "x$SENSU_BUILD_ITERATION" != "x" ]]; then
            echo $SENSU_BUILD_ITERATION
        else
            echo "Build iteration could not be found. If running locally you must set SENSU_BUILD_ITERATION."
            exit 1
        fi
    else
        # parse the iteration from the git tag
        TAG=$(git describe --exact-match --tags HEAD)
        RE="[0-9]+$"

        if [[ $TAG =~ $RE ]]; then
            echo ${BASH_REMATCH[0]}
        else
            echo "A build iteration could not be parsed from the tag: $TAG."
            exit 1
        fi
    fi
}

function usage() {
    echo "$script - output various version information"
    echo " "
    echo "$script [options]"
    echo " "
    echo "options:"
    echo "-h, --help                show help"
    echo "-t, --build-type          output the type of build this is"
    echo "-i, --iteration           output the iteration of the build"
    exit 0
}

case "$1" in
    -h|--help)
        usage
        ;;
    -t|--build-type)
        build_type
        ;;
    -i|--iteration)
        iteration
        ;;
    *)
        usage
        ;;
esac
