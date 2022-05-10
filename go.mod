module github.com/sensu/sensu-go

go 1.18

replace (
	github.com/sensu/sensu-go/api/core/v2 => ./api/core/v2
	github.com/sensu/sensu-go/api/core/v3 => ./api/core/v3
	github.com/sensu/sensu-go/backend/store/v2 => ./backend/store/v2
	github.com/sensu/sensu-go/types => ./types
)

require (
	github.com/AlecAivazis/survey/v2 v2.2.14
	github.com/atlassian/gostatsd v0.0.0-20180514010436-af796620006e
	github.com/blang/semver/v4 v4.0.0
	github.com/dave/jennifer v0.0.0-20171207062344-d8bdbdbee4e1
	github.com/docker/docker v0.0.0-20180409082103-cbde00b44273
	github.com/dustin/go-humanize v1.0.0
	github.com/echlebek/crock v1.0.1
	github.com/echlebek/migration v0.1.0
	github.com/echlebek/timeproxy v1.0.0
	github.com/emicklei/proto v1.1.0
	github.com/evanphx/json-patch/v5 v5.1.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-resty/resty/v2 v2.5.0
	github.com/gogo/protobuf v1.3.2
	github.com/golang-jwt/jwt/v4 v4.0.0
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.4
	github.com/google/uuid v1.1.2
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/websocket v1.4.2
	github.com/graph-gophers/dataloader v0.0.0-20180104184831-78139374585c
	github.com/graphql-go/graphql v0.7.10-0.20200426202700-116f19d099aa
	github.com/hashicorp/go-version v1.2.0
	github.com/influxdata/line-protocol v0.0.0-20210311194329-9aa0e372d097
	github.com/jackc/pgconn v1.12.1
	github.com/jackc/pgx/v4 v4.16.1
	github.com/json-iterator/go v1.1.12
	github.com/lib/pq v1.10.5
	github.com/mgutz/ansi v0.0.0-20170206155736-9520e82c474b
	github.com/mholt/archiver/v3 v3.3.1-0.20191129193105-44285f7ed244
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/hashstructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/olekukonko/tablewriter v0.0.5
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/client_model v0.2.0
	github.com/prometheus/common v0.26.0
	github.com/robertkrimen/otto v0.0.0-20191219234010-c382bd3c16ff
	github.com/robfig/cron/v3 v3.0.1
	github.com/sensu/lasr v1.2.1
	github.com/sensu/sensu-go/api/core/v2 v2.14.0
	github.com/sensu/sensu-go/api/core/v3 v3.6.1
	github.com/sensu/sensu-go/types v0.10.0
	github.com/shirou/gopsutil/v3 v3.21.12
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/willf/pad v0.0.0-20160331131008-b3d780601022
	go.etcd.io/bbolt v1.3.6
	go.etcd.io/etcd/api/v3 v3.5.2
	go.etcd.io/etcd/client/pkg/v3 v3.5.2
	go.etcd.io/etcd/client/v3 v3.5.2
	go.etcd.io/etcd/server/v3 v3.5.2
	go.etcd.io/etcd/tests/v3 v3.5.2
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.17.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
	google.golang.org/grpc v1.38.0
	gopkg.in/h2non/filetype.v1 v1.0.3
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/andybalholm/brotli v1.0.0 // indirect
	github.com/ash2k/stager v0.0.0-20170622123058-6e9c7b0eacd4 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/go-systemd/v22 v22.3.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/form3tech-oss/jwt-go v3.2.3+incompatible // indirect
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/fsnotify/fsnotify v1.4.7 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/gotestyourself/gotestyourself v2.2.0+incompatible // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	github.com/gxed/GoEndian v0.0.0-20160916112711-0f5c6873267e // indirect
	github.com/gxed/eventfd v0.0.0-20160916113412-80a92cca79a8 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/ipfs/go-log v1.0.4 // indirect
	github.com/ipfs/go-log/v2 v2.0.5 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/jackc/puddle v1.2.1 // indirect
	github.com/jbenet/go-reuseport v0.0.0-20180416043609-15a1cd37f050 // indirect
	github.com/jonboulle/clockwork v0.2.2 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/klauspost/compress v1.9.2 // indirect
	github.com/klauspost/pgzip v1.2.1 // indirect
	github.com/libp2p/go-reuseport v0.0.0-20180416043609-15a1cd37f050 // indirect
	github.com/libp2p/go-sockaddr v0.1.0 // indirect
	github.com/lufia/plan9stats v0.0.0-20211012122336-39d0f177ccd0 // indirect
	github.com/magiconair/properties v1.8.1 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mattn/go-isatty v0.0.12 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/opentracing/opentracing-go v1.1.0 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/pierrec/lz4/v3 v3.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/soheilhy/cmux v0.1.5 // indirect
	github.com/spf13/afero v1.1.2 // indirect
	github.com/spf13/cast v1.3.0 // indirect
	github.com/spf13/jwalterweatherman v1.0.0 // indirect
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	github.com/tklauser/numcpus v0.3.0 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20201229170055-e5319fda7802 // indirect
	github.com/ulikunitz/xz v0.5.10 // indirect
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	go.etcd.io/etcd/client/v2 v2.305.2 // indirect
	go.etcd.io/etcd/pkg/v3 v3.5.2 // indirect
	go.etcd.io/etcd/raft/v3 v3.5.2 // indirect
	go.opentelemetry.io/contrib v0.20.0 // indirect
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.20.0 // indirect
	go.opentelemetry.io/otel v0.20.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp v0.20.0 // indirect
	go.opentelemetry.io/otel/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v0.20.0 // indirect
	go.opentelemetry.io/otel/trace v0.20.0 // indirect
	go.opentelemetry.io/proto/otlp v0.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/genproto v0.0.0-20210602131652-f16073e35f0c // indirect
	google.golang.org/protobuf v1.26.0 // indirect
	gopkg.in/ini.v1 v1.51.0 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
	gopkg.in/sourcemap.v1 v1.0.5 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	gotest.tools v2.2.0+incompatible // indirect
	sigs.k8s.io/yaml v1.2.0 // indirect
)
