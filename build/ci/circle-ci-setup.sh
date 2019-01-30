#!/usr/bin/env bash

# This is largely a copy of sensu-go/Dockerfile.packaging, with some parts
# stripped out.
#
# It's intended to be temporary while we retool the packaging/release process
#

set -e
set -x

apt-get update
apt-get install -y build-essential git
apt-get clean
