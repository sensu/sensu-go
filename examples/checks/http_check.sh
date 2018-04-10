#!/usr/bin/env bash

target=$1

curl --silent --connect-timeout 5 $target > /dev/null
if [ $? -ne 0 ]; then
  exit 1
fi

exit 0
