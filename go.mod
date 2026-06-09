module github.com/sensu/sensu-go

go 1.26.3

replace github.com/gorilla/websocket => github.com/sensu/websocket v1.0.1

replace go.etcd.io/etcd/api/v3 => github.com/sensu/etcd/api/v3 v3.0.0

replace go.etcd.io/etcd/client/pkg/v3 => github.com/sensu/etcd/client/pkg/v3 v3.0.0

replace go.etcd.io/etcd/client/v3 => github.com/sensu/etcd/client/v3 v3.0.0

replace go.etcd.io/etcd/pkg/v3 => github.com/sensu/etcd/pkg/v3 v3.0.0

replace go.etcd.io/etcd/server/v3 => github.com/sensu/etcd/server/v3 v3.0.0

require (
	github.com/AlecAivazis/survey/v2 v2.2.14
	github.com/ash2k/stager v0.0.0-20170622123058-6e9c7b0eacd4 // indirect
	github.com/atlassian/gostatsd v0.0.0-20180514010436-af796620006e
	github.com/blang/semver/v4 v4.0.0
	github.com/dave/jennifer v1.7.1
	github.com/dustin/go-humanize v1.0.1
	github.com/echlebek/crock v1.0.1
	github.com/echlebek/timeproxy v1.0.0
	github.com/emicklei/proto v1.1.0
	github.com/evanphx/json-patch/v5 v5.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-resty/resty/v2 v2.11.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.5.2
	github.com/golang/protobuf v1.5.4
	github.com/golang/snappy v0.0.4
	github.com/google/uuid v1.6.0
	github.com/gorilla/mux v1.8.1
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
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.1.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/olekukonko/tablewriter v0.0.5
	github.com/prometheus/client_golang v1.20.5
	github.com/prometheus/client_model v0.6.1
	github.com/prometheus/common v0.62.0
	github.com/robertkrimen/otto v0.2.1
	github.com/robfig/cron/v3 v3.0.1
	github.com/sensu/lasr v1.2.1
	github.com/shirou/gopsutil/v3 v3.23.2
	github.com/sirupsen/logrus v1.9.3
	github.com/spf13/cobra v1.9.1
	github.com/spf13/pflag v1.0.10
	github.com/spf13/viper v1.21.0
	github.com/stretchr/testify v1.11.1
	github.com/willf/pad v0.0.0-20160331131008-b3d780601022
	go.etcd.io/bbolt v1.4.3
	go.etcd.io/etcd/api/v3 v3.6.8
	go.etcd.io/etcd/client/pkg/v3 v3.6.8
	go.etcd.io/etcd/client/v3 v3.6.8
	go.etcd.io/etcd/server/v3 v3.6.8
	go.uber.org/atomic v1.11.0
	go.uber.org/zap v1.28.0
	golang.org/x/crypto v0.51.0
	golang.org/x/mod v0.35.0
	golang.org/x/net v0.55.0
	golang.org/x/sys v0.45.0
	golang.org/x/time v0.14.0
	google.golang.org/grpc v1.81.1
	gopkg.in/h2non/filetype.v1 v1.0.3
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/coreos/go-systemd/v22 v22.5.0 // indirect
	github.com/creack/pty v1.1.20 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-test/deep v1.0.8
	github.com/google/btree v1.1.3 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.29.0 // indirect
	github.com/jonboulle/clockwork v0.5.0 // indirect
	github.com/kr/pty v1.1.8 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/sensu/core/v2 v2.21.2
	github.com/sensu/core/v3 v3.10.2
	github.com/sensu/sensu-api-tools v0.2.1
	github.com/sensu/sensu-go/types v0.13.0
	github.com/tmc/grpc-websocket-proxy v0.0.0-20220101234140-673ab2c3ae75 // indirect
	github.com/xiang90/probing v0.0.0-20221125231312-a49e3df8f510 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.59.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.44.0 // indirect
	go.uber.org/multierr v1.11.0
	google.golang.org/genproto/googleapis/api v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260526163538-3dc84a4a5aaa // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.2.1 // indirect
	sigs.k8s.io/yaml v1.6.0 // indirect
)

require golang.org/x/tools/go/vcs v0.1.0-deprecated

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v5 v5.0.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/coreos/go-semver v0.3.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fsnotify/fsnotify v1.10.1 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.0.1 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/ipfs/go-log/v2 v2.9.2 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.17.9 // indirect
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/pelletier/go-toml/v2 v2.3.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/tklauser/numcpus v0.6.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.etcd.io/etcd/pkg/v3 v3.6.8 // indirect
	go.etcd.io/raft/v3 v3.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.44.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.44.0 // indirect
	go.opentelemetry.io/otel/metric v1.44.0 // indirect
	go.opentelemetry.io/otel/sdk v1.44.0 // indirect
	go.opentelemetry.io/otel/trace v1.44.0 // indirect
	go.opentelemetry.io/proto/otlp v1.10.0 // indirect
	go.yaml.in/yaml/v2 v2.4.4 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/term v0.43.0 // indirect
	golang.org/x/text v0.37.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	sigs.k8s.io/json v0.0.0-20211020170558-c049b76a60c6 // indirect
)
