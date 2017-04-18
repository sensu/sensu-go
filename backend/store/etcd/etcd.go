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
	"net/url"
	"os"
	"path/filepath"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/pkg/capnslog"
)

const (
	// StateDir is the base path for Sensu's local storage.
	StateDir = "/var/lib/sensu"
	// EtcdStartupTimeout is the amount of time we give the embedded Etcd Server
	// to start.
	EtcdStartupTimeout = 60 // seconds
	// ClientListenURL is the default listen address for clients.
	ClientListenURL = "http://127.0.0.1:2379"
	// PeerListenURL is the default listen address for peers.
	PeerListenURL = "http://127.0.0.1:2380"
	// InitialCluster is the default initial cluster
	InitialCluster = "default=http://127.0.0.1:2380"
)

// Config is a configuration for the embedded etcd
type Config struct {
	StateDir        string
	PeerListenURL   string
	ClientListenURL string
	InitialCluster  string
}

// NewConfig returns a pointer to an initialized Config object with defaults.
func NewConfig() *Config {
	c := &Config{}
	c.StateDir = StateDir
	c.ClientListenURL = ClientListenURL
	c.PeerListenURL = PeerListenURL
	c.InitialCluster = InitialCluster

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

// NewEtcd returns a new, configured, and running Etcd. The running Etcd will
// panic on error. The calling goroutine should recover() from the panic and
// shutdown accordingly. Callers must also ensure that the running Etcd is
// cleanly shutdown before the process terminates.
//
// Callers should monitor the Err() channel for the running etcd--these are
// terminal errors.
func NewEtcd(config *Config) (*Etcd, error) {
	cfg := embed.NewConfig()
	cfgDir := filepath.Join(config.StateDir, "etcd", "data")
	walDir := filepath.Join(config.StateDir, "etcd", "wal")
	cfg.Dir = cfgDir
	cfg.WalDir = walDir
	if err := ensureDir(cfgDir); err != nil {
		return nil, err
	}
	if err := ensureDir(walDir); err != nil {
		return nil, err
	}

	clURL, err := url.Parse(config.ClientListenURL)
	if err != nil {
		return nil, err
	}

	apURL, err := url.Parse(config.PeerListenURL)
	if err != nil {
		return nil, err
	}

	cfg.ACUrls = []url.URL{*clURL}
	cfg.APUrls = []url.URL{*apURL}
	cfg.LCUrls = []url.URL{*clURL}
	cfg.LPUrls = []url.URL{*apURL}

	cfg.InitialCluster = config.InitialCluster

	capnslog.SetFormatter(NewLogrusFormatter())

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return nil, err
	}

	select {
	case <-e.Server.ReadyNotify():
		log.Info("Etcd ready to serve client connections")
	case <-time.After(EtcdStartupTimeout * time.Second):
		e.Server.Stop()
		return nil, fmt.Errorf("Etcd failed to start in %d seconds", EtcdStartupTimeout)
	}

	return &Etcd{config, e}, nil
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
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{e.cfg.ClientListenURL},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return cli, nil
}

// Healthy returns true if Etcd is healthy, false otherwise.
func (e *Etcd) Healthy() bool {
	client, err := e.NewClient()
	if err != nil {
		return false
	}
	mapi := clientv3.NewMaintenance(client)
	// TODO(greg): what can we do with the response? are there some operational
	// parameters that are useful?
	//
	// https://godoc.org/github.com/coreos/etcd/etcdserver/etcdserverpb#StatusResponse
	_, err = mapi.Status(context.TODO(), e.cfg.ClientListenURL)
	if err != nil {
		return false
	}
	return true
}
