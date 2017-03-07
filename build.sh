#!/usr/bin/env bash

REPO_PATH="github.com/sensu/sensu-go"

build_commands () {
  CGO_ENABLED=0 go build -o bin/sensu-agent ${REPO_PATH}/agent/cmd/...
}

build_commands
