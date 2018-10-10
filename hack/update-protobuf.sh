#!/usr/bin/env bash

ROOT=$(dirname "${BASH_SOURCE}")/..

PACKAGES=(
    github.com/sensu/sensu-go/apis/meta/v1alpha1
    github.com/sensu/sensu-go/apis/rbac/v1alpha1
)

go-to-protobuf \
    --proto-import=./vendor \
    --drop-embedded-fields=github.com/sensu/sensu-go/apis/meta/v1alpha1/meta.TypeMeta \
    --packages=$(IFS=, ; echo "${PACKAGES[*]}") \
    --go-header-file=${ROOT}/hack/update-protobuf-boilerplate.txt
