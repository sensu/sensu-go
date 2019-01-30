#!/usr/bin/env bash

# This is largely a copy of sensu-go/Dockerfile.packaging, with some parts
# stripped out.
#
# It's intended to be temporary while we retool the packaging/release process
#

set -e
set -x

apt-get update
apt-get install -y ruby ruby-dev build-essential rpm rpmlint wget git curl gcc
curl -sL https://deb.nodesource.com/setup_8.x | bash -
apt-get install -y nodejs
curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | apt-key add -
echo "deb https://dl.yarnpkg.com/debian/ stable main" | tee /etc/apt/sources.list.d/yarn.list
apt-get update
apt-get install yarn
# See https://github.com/jordansissel/fpm/issues/1592
gem install childprocess -v 0.9.0
gem install --no-ri --no-rdoc fpm
apt-get clean
