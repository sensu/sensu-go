#!/bin/sh

if [[ "x$TRAVIS_TAG" != "x" || $TRAVIS_EVENT_TYPE == "api" ]]; then
    echo "versioned"
    if [[ "x$TRAVIS_TAG" == "x" && $1 == "deploy_binaries" ]]; then
        echo "Nightlies do not support binary builds yet. Skipping."
        exit 0
    fi
    ./build.sh $1 versioned
elif [[ $TRAVIS_BRANCH == "master" ]]; then
    if [[ $1 == "deploy_docker" ]]; then
        ./build.sh $1 master
    else
        echo "./build.sh $1 is only run for tags & nightlies. Skipping."
    fi
else
    echo "Deployments not supported for this configuration. Skipping."
fi
