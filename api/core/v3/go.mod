module github.com/sensu/sensu-go/api/core/v3

go 1.13

replace github.com/sensu/sensu-go/api/core/v2 => ../v2

require (
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/sensu/sensu-go/api/core/v2 v2.0.0
)
