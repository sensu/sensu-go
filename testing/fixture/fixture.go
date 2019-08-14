package fixture

//go:generate -command protoc protoc --gofast_out=plugins:. -I=${GOPATH}/src:. -I=../../vendor/ -I=./ -I=../../vendor/github.com/gogo/protobuf/protobuf/
//go:generate protoc resource.proto
