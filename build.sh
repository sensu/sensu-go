#!/usr/bin/env bash

REPO_PATH="github.com/sensu/sensu-go"

cmd=${1:-"all"}

if [ "$GOARCH" == "amd64" ]; then
	RACE="--race"
fi

install_deps () {
  go get github.com/axw/gocov/gocov
  go get gopkg.in/alecthomas/gometalinter.v1
  go get github.com/gordonklaus/ineffassign
  go get github.com/jgautheron/goconst/cmd/goconst
  go get -u github.com/golang/lint/golint
}

build_commands () {
  echo "Running build..."

  go build -o bin/sensu-agent ${REPO_PATH}/agent/cmd/...
  go build -o bin/sensu-backend ${REPO_PATH}/backend/cmd/...
}

test_commands () {
  echo "Running tests..."

  gometalinter.v1 --vendor --disable-all --enable=vet --enable=vetshadow --enable=golint --enable=ineffassign --enable=goconst --tests ./...
  if [ $? -ne 0 ]; then
    echo "Linting failed..."
    exit 1
  fi

  go test -v $RACE $(go list ./... | egrep -v '(testing|vendor)')
  if [ $? -ne 0 ]; then
    echo "Tests failed..."
    exit 1
  fi
}

e2e_commands () {
  echo "Running e2e tests..."
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
