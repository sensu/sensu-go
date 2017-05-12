#!/usr/bin/env bash

target=$1

curl -s --connect-timeout 5 $target
if [ $? -ne 0 ]; then
  exit 1
fi

exit 0
