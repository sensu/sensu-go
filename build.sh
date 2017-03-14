#!/usr/bin/env bash

REPO_PATH="github.com/sensu/sensu-go"

cmd=${1:-"all"}

install_deps () {
  go get github.com/axw/gocov/gocov
  go get gopkg.in/alecthomas/gometalinter.v1
}

build_commands () {
  go build -o bin/sensu-agent ${REPO_PATH}/agent/cmd/...
  go build -o bin/sensu-backend ${REPO_PATH}/backend/cmd/...
}

test_commands () {
  gometalinter --vendor --disable-all --enable=vet --enable=vetshadow --enable=golint --enable=ineffassign --enable=goconst --tests ./...
  go test -v $(go list ./... | grep -v vendor)
}

if [ "$cmd" == "deps" ]; then
  install_deps
elif [ "$cmd" == "test" ]; then
  test_commands
elif [ "$cmd" == "build" ]; then
  build_commands
else
  install_deps
  test_commands
  build_commands
fi
