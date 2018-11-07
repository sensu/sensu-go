package types

//go:generate go run ../scripts/check_protoc/main.go
//go:generate go install github.com/sensu/sensu-go/vendor/github.com/gogo/protobuf/protoc-gen-gofast
//go:generate go install github.com/sensu/sensu-go/vendor/github.com/golang/protobuf/protoc-gen-go
//go:generate -command protoc protoc --gofast_out=plugins:. -I=../vendor/ -I=./ -I=../vendor/github.com/gogo/protobuf/protobuf/
//go:generate protoc adhoc.proto any.proto asset.proto authentication.proto check.proto entity.proto event.proto extension.proto filter.proto handler.proto hook.proto keepalive.proto meta.proto metrics.proto mutator.proto namespace.proto rbac.proto silenced.proto time_window.proto tls.proto user.proto
//go:generate go run ../scripts/make_typemap/make_typemap.go -t typemap.tmpl -o typemap.go
//go:generate go fmt typemap.go
