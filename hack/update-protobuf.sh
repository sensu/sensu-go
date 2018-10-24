#!/usr/bin/env bash

ROOT=$(dirname "${BASH_SOURCE}")/..

PACKAGES=(
    github.com/sensu/sensu-go/internal/apis/meta
    github.com/sensu/sensu-go/internal/apis/rbac
    github.com/sensu/sensu-go/internal/apis/core
    github.com/sensu/sensu-go/apis/meta/v1alpha1
)

go-to-protobuf \
    --proto-import="${ROOT}/vendor/github.com/gogo/protobuf/protobuf,${ROOT}/vendor" \
    --drop-embedded-fields=github.com/sensu/sensu-go/internal/apis/meta.TypeMeta \
    --packages=$(IFS=, ; echo "${PACKAGES[*]}") \
    --go-header-file=${ROOT}/hack/update-protobuf-boilerplate.txt \
    --apimachinery-packages="" \
    --output-base="$(go env GOPATH)/src"
