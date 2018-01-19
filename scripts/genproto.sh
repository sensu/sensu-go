#!/usr/bin/env bash
#
# Generate all protobuf bindings.
# Run from repository root.
#
# Following is largely borrowed from coreos/etcd/scripts/genproto.sh rev. 9abe9da
#
set -e

if ! [[ "$0" =~ scripts/genproto.sh ]]; then
    echo "must be run from repository root"
    exit 255
fi

PROTOC_VERSION="3.5.1"

if [[ $(protoc --version | cut -f2 -d' ') != "${PROTOC_VERSION}" ]]; then
    echo "could not find protoc ${PROTOC_VERSION}, is it installed + in PATH?"
    exit 255
fi

# directories containing protos to be built
DIRS="./types"

# exact version of packages to build
# TODO: possibly grab from Gopkg.toml
GOGO_PROTO_SHA="117892bf1866fbaa2318c03e50e40564c8845457"

# export paths
SENSU_ROOT="${GOPATH}/src/github.com/sensu/sensu-go"
GOGOPROTO_ROOT="${GOPATH}/src/github.com/gogo/protobuf"
GOGOPROTO_PATH="${GOGOPROTO_ROOT}:${GOGOPROTO_ROOT}/protobuf"

# install gogoproto dep
go get -u github.com/gogo/protobuf/{proto,protoc-gen-gogo,gogoproto}
go get -u golang.org/x/tools/cmd/goimports
pushd "${GOGOPROTO_ROOT}"
git reset --hard "${GOGO_PROTO_SHA}"
make install
popd

for dir in ${DIRS}; do
    pushd "$dir"
    protoc --gofast_out=plugins:. -I="$SENSU_ROOT/vendor/:./" ./*.proto
    goimports -w ./*.pb.go
    popd
done
