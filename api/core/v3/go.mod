module github.com/sensu/sensu-go/api/core/v3

go 1.13

replace (
	github.com/sensu/sensu-go/api/core/v2 => ../v2
	github.com/sensu/sensu-go/types => ../../../types
	go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20210226220824-aa7126864d82
)

require (
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/sensu/sensu-go/api/core/v2 v2.6.0
	github.com/sensu/sensu-go/types v0.3.0
)
