#!/usr/bin/env bash

# Build release for multiple platforms & archs and upload to GCS. Likely
# temporary until we can use Github releases when repo is public.

set -e

# GCS bucket where we'll deploy builds
BUCKET="sensu-binaries"
CMD=${1:-"all"}
TAG=""

install_deps() {
  pip install gsutil --user
}

# Checkout release tag
checkout_release () {
  TAG=$1 # Allow current tag to be overriden

  # If tag is specified switch to that branch
  if [$TAG != '']; then
    git checkout $TAG
  fi

  # Get the tag associated with the ref currently checked out
  TAG=$(git describe --tags)
  if [ $? -ne 0 ]; then
    echo 'Current ref must be associated with a tag'
    exit 1
  fi
}

clean_target_dir() {
  rm -rf ./target/*
}

build_supported_binaries() {
  clean_target_dir

  # ctl binaries
  GOOS=darwin GOARCH=amd64 ./build.sh build_cli
  GOOS=linux GOARCH=amd64 ./build.sh build_cli
  GOOS=linux GOARCH=386 ./build.sh build_cli
  GOOS=freebsd GOARCH=amd64 ./build.sh build_cli
  GOOS=windows GOARCH=amd64 ./build.sh build_cli
  GOOS=windows GOARCH=386 ./build.sh build_cli

  # backend binaries
  GOOS=darwin GOARCH=amd64 ./build.sh build_backend
  GOOS=linux GOARCH=amd64 ./build.sh build_backend
  GOOS=linux GOARCH=386 ./build.sh build_backend
  GOOS=freebsd GOARCH=amd64 ./build.sh build_backend
  GOOS=windows GOARCH=amd64 ./build.sh build_backend

  # agent binaries
  GOOS=darwin GOARCH=amd64 ./build.sh build_agent
  GOOS=linux GOARCH=amd64 ./build.sh build_agent
  GOOS=linux GOARCH=386 ./build.sh build_agent
  GOOS=freebsd GOARCH=amd64 ./build.sh build_agent
  GOOS=windows GOARCH=amd64 ./build.sh build_agent
  GOOS=windows GOARCH=386 ./build.sh build_agent
}

upload_supported_binaries() {
  upload_binary darwin amd64
  upload_binary freebsd amd64
  upload_binary linux 386
  upload_binary linux amd64
  upload_binary windows amd64
  upload_binary windows 386
}

upload_binary() {
  local os=$1
  local arch=$2

  gsutil -m cp -a public-read target/${os}-${arch}/* gs://${BUCKET}/${TAG}/${os}/${arch}/
  echo "Uploaded to https://storage.googleapis.com/${BUCKET}/${TAG}/${os}/${arch}/"
}

update_latest_txt() {
  echo $TAG > target/latest.txt
  gsutil cp -a public-read target/latest.txt gs://${BUCKET}/latest.txt
}

if [ "$CMD" == "all" ]; then
  checkout_release
  build_supported_binaries
  upload_supported_binaries
  update_latest_txt
elif [ "$CMD" == "deps" ]; then
  install_deps
# Build binaries for each platform & arch
elif [ "$CMD" == "release" ]; then
  checkout_release $1
  build_supported_binaries
  upload_supported_binaries
# Upload the file pointing to the latest release
elif [ "$CMD" == "set-latest" ]; then
  checkout_release $1
  update_latest_txt
fi
