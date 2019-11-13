package v2

//go:generate go run ../../../scripts/check_protoc/main.go
//go:generate -command protoc protoc --gofast_out=plugins:. -I=$GOPATH/pkg/mod -I=./ -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.2.1/protobuf
//go:generate protoc adhoc.proto any.proto apikey.proto asset.proto authentication.proto check.proto entity.proto event.proto extension.proto filter.proto handler.proto hook.proto keepalive.proto meta.proto metrics.proto mutator.proto namespace.proto rbac.proto silenced.proto tessen.proto time_window.proto tls.proto user.proto
//go:generate go run ../../../scripts/make_typemap/make_typemap.go -t typemap.tmpl -o typemap.go
//go:generate go fmt typemap.go
