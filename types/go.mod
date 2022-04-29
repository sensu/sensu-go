module github.com/sensu/sensu-go/types

go 1.13

replace (
	github.com/sensu/sensu-go => ../
	github.com/sensu/sensu-go/api/core/v2 => ../api/core/v2
	github.com/sensu/sensu-go/api/core/v3 => ../api/core/v3
)

require (
	github.com/blang/semver/v4 v4.0.0
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff
	github.com/sensu/sensu-go/api/core/v2 v2.14.0
	github.com/sensu/sensu-go/api/core/v3 v3.6.1
	github.com/stretchr/testify v1.6.0
)
