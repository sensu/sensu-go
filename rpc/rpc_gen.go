package rpc

//go:generate -command protoc protoc -I ../../../../ -I ../types/ -I . -I ../vendor/ --go_out=plugins=grpc:.
//go:generate protoc extension.proto
