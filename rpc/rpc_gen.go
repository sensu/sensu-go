package rpc

//go:generate go install github.com/sensu/sensu-go/vendor/github.com/golang/protobuf/protoc-gen-go
//go:generate -command protoc protoc -I ../../../../ -I . -I ../types/ -I ../vendor/ --go_out=plugins=grpc:.
//go:generate protoc extension.proto
