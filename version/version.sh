#!/bin/sh

# NOTE: make sure to run this from the sensu-go directory

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

# prerelease version will output the version of a prerelease from its tag
function prerelease_version() {
    b_type=$(build_type)
    case $b_type in
        alpha|beta|rc)
            TAG=$(git describe --exact-match --tags HEAD)
            RE=".*\-.*\.([0-9]+)\-[0-9]+$"

            if [[ $TAG =~ $RE ]]; then
                echo ${BASH_REMATCH[1]}
            else
                echo "A prerelease version could not be parsed from the tag: $TAG."
                exit 1
            fi
        ;;
        *)
            echo "Build type $b_type not supported by --prerelease-version."
            exit 1
            ;;
    esac
}

# version will output the version of the build (without iteration)
function version() {
    b_type=$(build_type)
    base_version=$(cat version/version.txt)
    case $b_type in
        nightly)
            echo $base_version-$b_type
            ;;
        alpha|beta|rc)
            echo $base_version-$b_type.$(prerelease_version)
            ;;
        stable)
            echo $base_version
    esac
}

# full_version will output the version of the build (with iteration)
function full_version() {
    echo $(version)-$(iteration)
}

function usage() {
    echo "$script - output various version information"
    echo " "
    echo "$script [options]"
    echo " "
    echo "options:"
    echo "-h, --help                show help"
    echo "-f, --full-version        output the version of the build with iteration"
    echo "-i, --iteration           output the iteration of the build"
    echo "-p, --prerelease-version  output the prerelease version of the build"
    echo "-t, --build-type          output the type of build this is"
    echo "-v, --version             output the version of the build without iteration"
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
    -p|--prerelease-version)
        prerelease_version
        ;;
    -v|--version)
        version
        ;;
    -f|--full-version)
        full_version
        ;;
    *)
        usage
        ;;
esac
