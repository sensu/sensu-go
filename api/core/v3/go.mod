module github.com/sensu/sensu-go/api/core/v3

go 1.13

replace (
	github.com/sensu/sensu-go/api/core/v2 => ../v2
	github.com/sensu/sensu-go/types => ../../../types
)

require (
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.5
	github.com/sensu/sensu-go/api/core/v2 v2.6.0
	github.com/sensu/sensu-go/types v0.3.0
)
