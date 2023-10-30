// Package etcd manages the embedded etcd server that Sensu uses for storing
// state consistently across sensu-backend processes.
//
// To use the embedded Etcd, you must first call NewEtcd(). This will configure
// and start Etcd and ensure that it starts correctly. The channel returned by
// Err() should be monitored--these are terminal/fatal errors for Etcd.
package etcd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/sensu/sensu-go/util/path"
	"go.etcd.io/etcd/client/pkg/v3/logutil"
	"go.etcd.io/etcd/client/pkg/v3/transport"
	etcdTypes "go.etcd.io/etcd/client/pkg/v3/types"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"go.etcd.io/etcd/server/v3/etcdserver/api/v3rpc"
	"go.etcd.io/etcd/server/v3/proxy/grpcproxy/adapter"
	zapcore "go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

const (
	// ClusterStateNew specifies this is a new etcd cluster
	ClusterStateNew = "new"
	// EtcdStartupTimeout is the amount of time we give the embedded Etcd Server
	// to start.
	EtcdStartupTimeout = 60 // seconds

	// DefaultMaxRequestBytes is the default maximum request size for etcd
	// requests (1.5 MB)
	DefaultMaxRequestBytes = 1.5 * (1 << 20)

	// DefaultQuotaBackendBytes is the default database size limit for etcd
	// databases (4 GB)
	DefaultQuotaBackendBytes int64 = (1 << 32)

	// DefaultTickMs is the default Heartbeat Interval. This is the interval
	// with which the leader will notify followers that it is still the leader.
	// For best practices, the parameter should be set around round-trip time
	// between members.
	// See: https://github.com/etcd-io/etcd/blob/master/Documentation/tuning.md#time-parameters
	DefaultTickMs = 300

	// DefaultElectionMs is the default Election Timeout. This timeout is how
	// long a follower node will go without hearing a heartbeat before
	// attempting to become leader itself.
	// See: https://github.com/etcd-io/etcd/blob/master/Documentation/tuning.md#time-parameters
	DefaultElectionMs = 3000

	// DefaultLogLevel is the default log level for the embedded etcd server.
	DefaultLogLevel = "warn"

	// DefaultClientLogLevel is the default log level for the etcd client.
	DefaultClientLogLevel = "error"
)

func init() {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))

	logutil.DefaultZapLoggerConfig.Encoding = "sensu-json"
	logutil.DefaultZapLoggerConfig.EncoderConfig.TimeKey = "time"
	logutil.DefaultZapLoggerConfig.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	logutil.DefaultZapLoggerConfig.EncoderConfig.EncodeLevel = sensuLevelEncoder
}

// Config is a configuration for the embedded etcd
type Config struct {
	DataDir string

	// Cluster Member Name
	Name string

	// Heartbeat interval
	TickMs uint

	// Election Timeout
	ElectionMs uint

	AdvertiseClientURLs      []string
	ListenPeerURLs           []string
	ListenClientURLs         []string
	InitialCluster           string
	InitialClusterState      string
	InitialClusterToken      string
	InitialAdvertisePeerURLs []string
	Discovery                string
	DiscoverySrv             string

	ClientTLSInfo TLSInfo
	PeerTLSInfo   TLSInfo

	CipherSuites []string

	MaxRequestBytes   uint
	QuotaBackendBytes int64

	LogLevel           string
	ClientLogLevel     string
	LogTimestampLayout string

	UnsafeNoFsync bool
}

// TLSInfo wraps etcd transport TLSInfo
type TLSInfo transport.TLSInfo

// NewConfig returns a pointer to an initialized Config object with defaults.
func NewConfig() *Config {
	c := &Config{}
	c.DataDir = path.SystemCacheDir("sensu-backend")
	c.MaxRequestBytes = DefaultMaxRequestBytes
	c.QuotaBackendBytes = DefaultQuotaBackendBytes
	c.TickMs = DefaultTickMs
	c.ElectionMs = DefaultElectionMs
	c.LogLevel = DefaultLogLevel
	c.ClientLogLevel = DefaultClientLogLevel

	return c
}

func ensureDir(path string) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if mkdirErr := os.MkdirAll(path, 0700); mkdirErr != nil {
				return mkdirErr
			}
		} else {
			return err
		}
	}
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("path exists and is not directory - %s", path)
	}
	return nil
}

// Etcd is a wrapper around go.etcd.io/etcd/embed.Etcd
type Etcd struct {
	cfg  *Config
	etcd *embed.Etcd
}

// BackendID returns the ID of the etcd cluster member
func (e *Etcd) BackendID() (result string) {
	return e.etcd.Server.ID().String()
}

// GetClusterVersion returns the cluster version of the etcd server
func (e *Etcd) GetClusterVersion() string {
	version := e.etcd.Server.ClusterVersion()
	if version == nil {
		logger.Error("nil cluster version!")
		return "UNKNOWN"
	}
	return version.String()
}

// NewEtcd returns a new, configured, and running Etcd. The running Etcd will
// panic on error. The calling goroutine should recover() from the panic and
// shutdown accordingly. Callers must also ensure that the running Etcd is
// cleanly shutdown before the process terminates.
//
// Callers should monitor the Err() channel for the running etcd--these are
// terminal errors.
func NewEtcd(config *Config) (*Etcd, error) {
	// Parse the various URLs
	var err error
	var lcURLs etcdTypes.URLs
	lcURLs, err = etcdTypes.NewURLs(config.ListenClientURLs)
	if err != nil {
		return nil, fmt.Errorf("invalid listen client urls: %s", err)
	}
	var acURLs etcdTypes.URLs
	acURLs, err = etcdTypes.NewURLs(config.AdvertiseClientURLs)
	if err != nil {
		return nil, fmt.Errorf("invalid advertise client urls: %s", err)
	}
	var lpURLs etcdTypes.URLs
	lpURLs, err = etcdTypes.NewURLs(config.ListenPeerURLs)
	if err != nil {
		return nil, fmt.Errorf("invalid listen peer urls: %s", err)
	}
	var apURLs etcdTypes.URLs
	apURLs, err = etcdTypes.NewURLs(config.InitialAdvertisePeerURLs)
	if err != nil {
		return nil, fmt.Errorf("invalid initial advertise peer urls: %s", err)
	}

	cfg := embed.NewConfig()
	if config.Name != "" {
		cfg.Name = config.Name
	}

	cfg.Dir = filepath.Join(config.DataDir, "etcd", "data")
	if err := ensureDir(cfg.Dir); err != nil {
		return nil, err
	}

	cfg.WalDir = filepath.Join(config.DataDir, "etcd", "wal")
	if err := ensureDir(cfg.WalDir); err != nil {
		return nil, err
	}

	// Heartbeat and Election timeouts
	if config.TickMs != 0 {
		cfg.TickMs = config.TickMs
	}

	if config.ElectionMs != 0 {
		cfg.ElectionMs = config.ElectionMs
	}

	// Client config
	cfg.AdvertiseClientUrls = acURLs
	cfg.ListenClientUrls = lcURLs
	cfg.ClientTLSInfo = (transport.TLSInfo)(config.ClientTLSInfo)

	// Peer config
	cfg.AdvertisePeerUrls = apURLs
	cfg.ListenPeerUrls = lpURLs
	cfg.PeerTLSInfo = (transport.TLSInfo)(config.PeerTLSInfo)

	if len(config.CipherSuites) > 0 {
		cfg.CipherSuites = config.CipherSuites
	}

	// Cluster config
	if config.InitialClusterToken != "" {
		cfg.InitialClusterToken = config.InitialClusterToken
	}

	if config.InitialCluster != "" {
		cfg.InitialCluster = config.InitialCluster
	}

	if config.InitialClusterState != "" {
		cfg.ClusterState = config.InitialClusterState
	}

	if config.Discovery != "" {
		cfg.Durl = config.Discovery
	}

	if config.DiscoverySrv != "" {
		cfg.DNSCluster = config.DiscoverySrv
	}

	// Every 5 minutes, we will prune all values in etcd to only their latest
	// revision.
	cfg.AutoCompactionMode = "revision"
	cfg.AutoCompactionRetention = "2"

	if config.QuotaBackendBytes != 0 {
		cfg.QuotaBackendBytes = config.QuotaBackendBytes
	}

	if config.MaxRequestBytes != 0 {
		cfg.MaxRequestBytes = config.MaxRequestBytes
	}

	cfg.Logger = "zap"

	if config.LogLevel != "" {
		cfg.LogLevel = config.LogLevel
		logutil.DefaultZapLoggerConfig.Level.SetLevel(LogLevelToZap(config.LogLevel))
	}

	if config.LogTimestampLayout != "" {
		encoder := zapcore.TimeEncoderOfLayout(config.LogTimestampLayout)
		logutil.DefaultZapLoggerConfig.EncoderConfig.EncodeTime = encoder
	}

	// Unsafe options.
	cfg.UnsafeNoFsync = config.UnsafeNoFsync

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}

	select {
	case <-e.Server.ReadyNotify():
		logger.Info("Etcd ready to serve client connections")
	case <-time.After(EtcdStartupTimeout * time.Second):
		e.Server.Stop()
		return nil, fmt.Errorf("Etcd failed to start in %d seconds", EtcdStartupTimeout)
	}

	return &Etcd{cfg: config, etcd: e}, nil
}

// Name returns the configured name for Etcd.
func (e *Etcd) Name() string {
	return e.cfg.Name
}

// Err returns the error channel for Etcd or nil if no etcd is started.
func (e *Etcd) Err() <-chan error {
	return e.etcd.Err()
}

// Shutdown will cleanly shutdown the running Etcd.
func (e *Etcd) Shutdown() error {
	etcdStopped := e.etcd.Server.StopNotify()
	e.etcd.Close()
	<-etcdStopped
	return nil
}

// NewClient returns a new etcd v3 client.
func (e *Etcd) NewClient() (*clientv3.Client, error) {
	return e.NewClientContext(context.Background())
}

// NewClientContext is like NewClient, but sets the provided context on the
// client.
func (e *Etcd) NewClientContext(ctx context.Context) (*clientv3.Client, error) {
	tlsConfig, err := ((transport.TLSInfo)(e.cfg.ClientTLSInfo)).ClientConfig()
	if err != nil {
		return nil, err
	}

	// Set etcd client log level
	logConfig := logutil.DefaultZapLoggerConfig
	logConfig.Level.SetLevel(LogLevelToZap(e.cfg.ClientLogLevel))

	return clientv3.New(clientv3.Config{
		Endpoints:   e.cfg.AdvertiseClientURLs,
		DialTimeout: 60 * time.Second,
		TLS:         tlsConfig,
		DialOptions: []grpc.DialOption{
			grpc.WithReturnConnectionError(),
			grpc.WithBlock(),
		},
		Context:   ctx,
		LogConfig: &logConfig,
	})
}

// NewEmbeddedClient delivers a new embedded etcd client. Only for testing.
func (e *Etcd) NewEmbeddedClient() *clientv3.Client {
	return e.NewEmbeddedClientWithContext(context.Background())
}

// NewEmbeddedClientWithContext takes a context and delivers a new embedded etcd
// client. Only for testing.
// Based on https://github.com/etcd-io/etcd/blob/v3.4.16/etcdserver/api/v3client/v3client.go#L30.
func (e *Etcd) NewEmbeddedClientWithContext(ctx context.Context) *clientv3.Client {
	// Set etcd client log level
	logConfig := logutil.DefaultZapLoggerConfig
	logConfig.Level.SetLevel(LogLevelToZap(e.cfg.ClientLogLevel))
	clientLogger, err := logConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("error building etcd client logger: %s", err))
	}

	c := clientv3.NewCtxClient(ctx, clientv3.WithZapLogger(clientLogger))

	kvc := adapter.KvServerToKvClient(v3rpc.NewQuotaKVServer(e.etcd.Server))
	c.KV = clientv3.NewKVFromKVClient(kvc, c)

	lc := adapter.LeaseServerToLeaseClient(v3rpc.NewQuotaLeaseServer(e.etcd.Server))
	c.Lease = clientv3.NewLeaseFromLeaseClient(lc, c, time.Second)

	wc := adapter.WatchServerToWatchClient(v3rpc.NewWatchServer(e.etcd.Server))
	c.Watcher = &watchWrapper{clientv3.NewWatchFromWatchClient(wc, c)}

	mc := adapter.MaintenanceServerToMaintenanceClient(v3rpc.NewMaintenanceServer(e.etcd.Server))
	c.Maintenance = clientv3.NewMaintenanceFromMaintenanceClient(mc, c)

	clc := adapter.ClusterServerToClusterClient(v3rpc.NewClusterServer(e.etcd.Server))
	c.Cluster = clientv3.NewClusterFromClusterClient(clc, c)

	return c
}

// Healthy returns Etcd status information. DEPRECATED.
func (e *Etcd) Healthy() bool {
	if len(e.cfg.AdvertiseClientURLs) == 0 {
		return false
	}
	client := e.NewEmbeddedClient()
	mapi := client.Maintenance
	_, err := mapi.Status(context.TODO(), e.cfg.AdvertiseClientURLs[0])
	return err == nil
}

// GetClientURLs gets the valid client URLs for the etcd server.
func (e *Etcd) GetClientURLs() []string {
	listeners := e.etcd.Clients
	results := make([]string, 0, len(listeners))
	for _, list := range listeners {
		results = append(results, list.Addr().String())
	}
	return results
}

// BlankContext implements Stringer on a context so the ctx string doesn't
// depend on the context's WithValue data, which tends to be unsynchronized
// (e.g., x/net/trace), causing ctx.String() to throw data races.
type blankContext struct{ context.Context }

func (*blankContext) String() string { return "(blankCtx)" }

// watchWrapper wraps clientv3 watch calls to blank out the context
// to avoid races on trace data.
type watchWrapper struct{ clientv3.Watcher }

func (ww *watchWrapper) Watch(ctx context.Context, key string, opts ...clientv3.OpOption) clientv3.WatchChan {
	return ww.Watcher.Watch(&blankContext{ctx}, key, opts...)
}

func LogLevelToZap(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "dpanic":
		return zapcore.DPanicLevel
	case "panic":
		return zapcore.PanicLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		panic(fmt.Sprintf("invalid etcd log level: %s", level))
	}
}
