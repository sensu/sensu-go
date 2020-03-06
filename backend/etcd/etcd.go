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

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/etcd/pkg/transport"
	etcdTypes "github.com/coreos/etcd/pkg/types"
	"github.com/coreos/pkg/capnslog"
	"github.com/sensu/sensu-go/util/path"
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
	DefaultTickMs = 100

	// DefaultElectionMs is the default Election Timeout. This timeout is how
	// long a follower node will go without hearing a heartbeat before
	// attempting to become leader itself.
	// See: https://github.com/etcd-io/etcd/blob/master/Documentation/tuning.md#time-parameters
	DefaultElectionMs = 1000
)

func init() {
	clientv3.SetLogger(grpclog.NewLoggerV2(ioutil.Discard, ioutil.Discard, ioutil.Discard))
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

// Etcd is a wrapper around github.com/coreos/etcd/embed.Etcd
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
	cfg.Name = config.Name

	cfg.Dir = filepath.Join(config.DataDir, "etcd", "data")
	cfg.WalDir = filepath.Join(config.DataDir, "etcd", "wal")
	if err := ensureDir(cfg.Dir); err != nil {
		return nil, err
	}
	if err := ensureDir(cfg.WalDir); err != nil {
		return nil, err
	}

	// Heartbeat and Election timeouts
	cfg.TickMs = config.TickMs
	cfg.ElectionMs = config.ElectionMs

	// Client config
	cfg.ACUrls = acURLs
	cfg.LCUrls = lcURLs
	cfg.ClientTLSInfo = (transport.TLSInfo)(config.ClientTLSInfo)

	// Peer config
	cfg.APUrls = apURLs
	cfg.LPUrls = lpURLs
	cfg.PeerTLSInfo = (transport.TLSInfo)(config.PeerTLSInfo)

	cfg.CipherSuites = config.CipherSuites

	// Cluster config
	cfg.InitialClusterToken = config.InitialClusterToken
	cfg.InitialCluster = config.InitialCluster
	cfg.ClusterState = config.InitialClusterState
	cfg.Durl = config.Discovery
	cfg.DNSCluster = config.DiscoverySrv

	// Every 5 minutes, we will prune all values in etcd to only their latest
	// revision.
	cfg.AutoCompactionMode = "revision"
	cfg.AutoCompactionRetention = "2"
	cfg.QuotaBackendBytes = config.QuotaBackendBytes
	cfg.MaxRequestBytes = config.MaxRequestBytes

	capnslog.SetFormatter(NewLogrusFormatter())

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

// NewClient returns a new etcd v3 client. Clients must be closed after use.
func (e *Etcd) NewClient() (*clientv3.Client, error) {
	tlsConfig, err := ((transport.TLSInfo)(e.cfg.ClientTLSInfo)).ClientConfig()
	if err != nil {
		return nil, err
	}
	return clientv3.New(clientv3.Config{
		Endpoints:   e.cfg.AdvertiseClientURLs,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
		DialOptions: []grpc.DialOption{
			grpc.WithBlock(),
		},
	})
}

// Healthy returns Etcd status information.
func (e *Etcd) Healthy() bool {
	if len(e.cfg.AdvertiseClientURLs) == 0 {
		return false
	}
	client, err := e.NewClient()
	if err != nil {
		return false
	}
	mapi := client.Maintenance
	_, err = mapi.Status(context.TODO(), e.cfg.AdvertiseClientURLs[0])
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
