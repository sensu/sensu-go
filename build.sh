#!/usr/bin/env bash

REPO_PATH="github.com/sensu/sensu-go"

build_commands () {
  go build -o bin/sensu-agent ${REPO_PATH}/agent/cmd/...
  go build -o bin/sensu-backend ${REPO_PATH}/backend/cmd/...
}

build_commands
