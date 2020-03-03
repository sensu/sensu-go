package v2

// TODO: go build will build the latest version of the protoc-gen-gofast module
// used by the project. We need a way to pin the exact version of
// protoc-gen-gofast used in this generator. We could not find a way of doing
// this at the time of this writing.

//go:generate go run ../../../scripts/check_protoc/main.go
//go:generate go build -o $GOPATH/bin/protoc-gen-gofast github.com/gogo/protobuf/protoc-gen-gofast
//go:generate -command protoc protoc --plugin $GOPATH/bin/protoc-gen-gofast --gofast_out=plugins:. -I=$GOPATH/pkg/mod -I=./ -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.1/protobuf
//go:generate protoc adhoc.proto any.proto apikey.proto asset.proto authentication.proto check.proto entity.proto event.proto extension.proto filter.proto handler.proto hook.proto keepalive.proto meta.proto metrics.proto mutator.proto namespace.proto rbac.proto secret.proto silenced.proto tessen.proto time_window.proto tls.proto user.proto
//go:generate go run ../../../scripts/make_typemap/make_typemap.go -t typemap.tmpl -o typemap.go
//go:generate go fmt typemap.go
