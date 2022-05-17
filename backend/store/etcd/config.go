package etcd

import "github.com/sensu/sensu-go/backend/etcd"

type Config struct {
	ClientTLSInfo     etcd.TLSInfo
	URLs              []string
	Username          string
	Password          string
	LogLevel          string
	UseEmbeddedClient bool
}
