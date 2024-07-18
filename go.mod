module github.com/sensu/sensu-go

go 1.22

require (
	github.com/AlecAivazis/survey/v2 v2.2.14
	github.com/ash2k/stager v0.0.0-20170622123058-6e9c7b0eacd4 // indirect
	github.com/atlassian/gostatsd v0.0.0-20180514010436-af796620006e
	github.com/blang/semver/v4 v4.0.0
	github.com/dave/jennifer v0.0.0-20171207062344-d8bdbdbee4e1
	github.com/dustin/go-humanize v1.0.1
	github.com/echlebek/crock v1.0.1
	github.com/echlebek/timeproxy v1.0.0
	github.com/emicklei/proto v1.1.0
	github.com/evanphx/json-patch/v5 v5.1.0
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/ghodss/yaml v1.0.0
	github.com/go-resty/resty/v2 v2.5.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/golang/protobuf v1.5.3
	github.com/golang/snappy v0.0.4
	github.com/google/uuid v1.4.0
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.5.0
	github.com/graph-gophers/dataloader v0.0.0-20180104184831-78139374585c
	github.com/graphql-go/graphql v0.8.1
	github.com/gxed/GoEndian v0.0.0-20160916112711-0f5c6873267e // indirect
	github.com/gxed/eventfd v0.0.0-20160916113412-80a92cca79a8 // indirect
	github.com/hashicorp/go-version v1.2.0
	github.com/influxdata/line-protocol v0.0.0-20210311194329-9aa0e372d097
	github.com/ipfs/go-log v1.0.4 // indirect
	github.com/jbenet/go-reuseport v0.0.0-20180416043609-15a1cd37f050 // indirect
	github.com/libp2p/go-reuseport v0.0.0-20180416043609-15a1cd37f050 // indirect
	github.com/libp2p/go-sockaddr v0.1.0 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mholt/archiver/v3 v3.3.1-0.20191129193105-44285f7ed244
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/olekukonko/tablewriter v0.0.5
	github.com/prometheus/client_golang v1.17.0
	github.com/prometheus/client_model v0.5.0
	github.com/prometheus/common v0.45.0
	github.com/robertkrimen/otto v0.2.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/sensu/lasr v1.2.1
	github.com/shirou/gopsutil/v3 v3.23.2
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.8.4
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/willf/pad v0.0.0-20160331131008-b3d780601022
	go.etcd.io/bbolt v1.3.8
	go.etcd.io/etcd/api/v3 v3.5.10
	go.etcd.io/etcd/client/pkg/v3 v3.5.10
	go.etcd.io/etcd/client/v3 v3.5.10
	go.etcd.io/etcd/server/v3 v3.5.10
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.26.0
	golang.org/x/crypto v0.15.0
	golang.org/x/mod v0.11.0
	golang.org/x/net v0.18.0
	golang.org/x/sys v0.14.0
	golang.org/x/time v0.3.0
	golang.org/x/tools v0.10.0
	google.golang.org/grpc v1.59.0
	gopkg.in/h2non/filetype.v1 v1.0.3
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/coreos/go-systemd v0.0.0-20190719114852-fd7a80b32e1f // indirect
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/creack/pty v1.1.20 // indirect
	github.com/go-logr/logr v1.3.0 // indirect
	github.com/go-test/deep v1.0.8
	github.com/google/btree v1.1.2 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.18.0 // indirect
	github.com/jonboulle/clockwork v0.4.0 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/sensu/core/v2 v2.20.0
	github.com/sensu/core/v3 v3.9.0
	github.com/sensu/sensu-api-tools v0.2.1
	github.com/sensu/sensu-go/types v0.13.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20220101234140-673ab2c3ae75 // indirect
	github.com/xiang90/probing v0.0.0-20221125231312-a49e3df8f510 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.45.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	google.golang.org/genproto v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20231120223509-83a465c0220f // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231120223509-83a465c0220f // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	sigs.k8s.io/yaml v1.4.0 // indirect
)
