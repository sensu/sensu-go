package v3

// TODO: go build will build the latest version of the protoc-gen-gofast module
// used by the project. We need a way to pin the exact version of
// protoc-gen-gofast used in this generator. We could not find a way of doing
// this at the time of this writing.

//go:generate go run ./internal/codegen/check_protoc
//go:generate go build -o $GOPATH/bin/protoc-gen-gofast github.com/gogo/protobuf/protoc-gen-gofast
//go:generate -command protoc protoc --plugin $GOPATH/bin/protoc-gen-gofast --gofast_out=plugins:$GOPATH/src -I=$GOPATH/pkg/mod -I=$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.1/protobuf -I=$GOPATH/src
//go:generate protoc github.com/sensu/sensu-go/api/core/v3/entity_state.proto github.com/sensu/sensu-go/api/core/v3/entity_config.proto
//go:generate protoc github.com/sensu/sensu-go/api/core/v3/pipeline.proto
//go:generate go run ./internal/codegen/generate_type -t typemap.tmpl -o typemap.go
//go:generate go fmt typemap.go
//go:generate go run ./internal/codegen/generate_type -t typemap_test.tmpl -o typemap_test.go
//go:generate go fmt typemap_test.go
//go:generate go run ./internal/codegen/generate_type -t resource_generated.tmpl -o resource_generated.go
//go:generate go fmt resource_generated.go
//go:generate go run ./internal/codegen/generate_type -t resource_generated_test.tmpl -o resource_generated_test.go
//go:generate go fmt resource_generated_test.go
