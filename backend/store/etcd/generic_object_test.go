package etcd

//go:generate -command protoc protoc --gofast_out=plugins:. -I=../../../vendor/ -I=./  -I=../../../api/core/v2/ -I=../../../api/core/v2/meta.proto -I=../../../vendor/github.com/gogo/protobuf/protobuf/
//go:generate protoc generic_object.proto
