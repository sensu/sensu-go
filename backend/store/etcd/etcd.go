// Package etcd manages the embedded etcd server that Sensu uses for storing
// state consistently across sensu-backend processes.
//
// To use the embedded Etcd, you must first call NewEtcd(). This will configure
// and start Etcd and ensure that it starts correctly. The goroutine monitoring
// Etcd for a fatal error will cause a panic should Etcd fail. The calling
// goroutine should recover() from the panic and shutdown appropriately.
package etcd

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
)

const (
	// StateDir is the base path for Sensu's local storage.
	StateDir = "/var/cache/sensu"
	// EtcdStartupTimeout is the amount of time we give the embedded Etcd Server
	// to start.
	EtcdStartupTimeout = 60 // seconds

	etcdClientAddress = "127.0.0.1:2379"
)

var (
	// the running etcd is a private singleton and we just act on it with
	// functions in here.
	etcdServer *embed.Etcd
)

// NewEtcd returns a new, configured, and running Etcd. The running Etcd will
// panic on error. The calling goroutine should recover() from the panic and
// shutdown accordingly. Callers must also ensure that the running Etcd is
// cleanly shutdown before the process terminates.
func NewEtcd() error {
	cfg := embed.NewConfig()
	cfg.Dir = filepath.Join(StateDir, "etcd", "data")
	cfg.WalDir = filepath.Join(StateDir, "etcd", "wal")

	e, err := embed.StartEtcd(cfg)
	if err != nil {
		return err
	}

	select {
	case <-e.Server.ReadyNotify():
		log.Println("Etcd ready to serve client connections")
	case <-time.After(EtcdStartupTimeout * time.Second):
		e.Server.Stop()
		return fmt.Errorf("Etcd failed to start in %d seconds", EtcdStartupTimeout)
	}

	go func() {
		log.Fatal(<-e.Err())
	}()

	etcdServer = e

	return nil
}

// Shutdown will cleanly shutdown the running Etcd.
func Shutdown() error {
	if etcdServer == nil {
		return errors.New("no running etcd detected")
	}

	etcdServer.Server.Stop()
	etcdServer.Close()
	return nil
}

// NewClient returns a new etcd v3 client. Clients must be closed after use.
func NewClient() (*clientv3.Client, error) {
	if etcdServer == nil {
		return nil, errors.New("no running etcd - must call NewEtcd() prior to NewClient()")
	}

	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdClientAddress},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	return cli, nil
}
