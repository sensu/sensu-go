module github.com/sensu/sensu-go/api/core/v3

go 1.13

replace (
	github.com/sensu/sensu-go/api/core/v2 => ../v2
	github.com/sensu/sensu-go/types => ../../../types
)

require (
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.0.0 // indirect
	github.com/golang/protobuf v1.5.2
	github.com/sensu/sensu-go/api/core/v2 v2.6.0
	github.com/sensu/sensu-go/types v0.3.0
)
