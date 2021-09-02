package v2

// TODO: go build will build the latest version of the protoc-gen-gofast module
// used by the project. We need a way to pin the exact version of
// protoc-gen-gofast used in this generator. We could not find a way of doing
// this at the time of this writing.

//go:generate go run ./internal/codegen/check_protoc
//go:generate go build -o $GOPATH/bin/protoc-gen-gofast github.com/gogo/protobuf/protoc-gen-gofast
//go:generate -command protoc protoc --plugin $GOPATH/bin/protoc-gen-gofast --gofast_out=plugins:$GOPATH/src -I=$GOPATH/pkg/mod -I=$GOPATH/src -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.1/protobuf
//go:generate protoc github.com/sensu/sensu-go/api/core/v2/adhoc.proto github.com/sensu/sensu-go/api/core/v2/any.proto github.com/sensu/sensu-go/api/core/v2/apikey.proto github.com/sensu/sensu-go/api/core/v2/asset.proto github.com/sensu/sensu-go/api/core/v2/authentication.proto github.com/sensu/sensu-go/api/core/v2/check.proto github.com/sensu/sensu-go/api/core/v2/entity.proto github.com/sensu/sensu-go/api/core/v2/event.proto github.com/sensu/sensu-go/api/core/v2/filter.proto github.com/sensu/sensu-go/api/core/v2/handler.proto github.com/sensu/sensu-go/api/core/v2/hook.proto github.com/sensu/sensu-go/api/core/v2/keepalive.proto github.com/sensu/sensu-go/api/core/v2/meta.proto github.com/sensu/sensu-go/api/core/v2/metrics.proto github.com/sensu/sensu-go/api/core/v2/mutator.proto github.com/sensu/sensu-go/api/core/v2/namespace.proto github.com/sensu/sensu-go/api/core/v2/rbac.proto github.com/sensu/sensu-go/api/core/v2/secret.proto github.com/sensu/sensu-go/api/core/v2/silenced.proto github.com/sensu/sensu-go/api/core/v2/tessen.proto github.com/sensu/sensu-go/api/core/v2/time_window.proto github.com/sensu/sensu-go/api/core/v2/tls.proto github.com/sensu/sensu-go/api/core/v2/user.proto
//go:generate protoc github.com/sensu/sensu-go/api/core/v2/pipeline.proto github.com/sensu/sensu-go/api/core/v2/pipeline_workflow.proto github.com/sensu/sensu-go/api/core/v2/resource_reference.proto
//go:generate go run ./internal/codegen/generate_type -t typemap.tmpl -o typemap.go
//go:generate go fmt typemap.go
//go:generate go run ./internal/codegen/generate_type -t typemap_test.tmpl -o typemap_test.go
//go:generate go fmt typemap_test.go
