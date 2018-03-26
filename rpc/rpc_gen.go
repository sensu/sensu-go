package rpc

//go:generate -command protoc protoc -I ../../../../ -I . -I ../types/ -I ../vendor/ --go_out=plugins=grpc:.
//go:generate protoc extension.proto
