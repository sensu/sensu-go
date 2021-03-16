module github.com/sensu/sensu-go/api/core/v2

go 1.13

replace go.etcd.io/etcd => go.etcd.io/etcd v0.0.0-20210226220824-aa7126864d82

require (
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/echlebek/crock v1.0.1
	github.com/echlebek/timeproxy v1.0.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/json-iterator/go v1.1.9
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.0
	go.etcd.io/etcd v3.4.15+incompatible
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)
