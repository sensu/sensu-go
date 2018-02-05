package leader

import (
	"errors"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

var (
	// sensuLeaderKey is to be used in concert with KeyBuilder
	sensuLeaderKey = "/leader/"
)

var (
	// ErrNotInitialized is returned when Init has not been called.
	ErrNotInitialized = errors.New("package not initialized")
)

var (
	pkgMu    sync.Mutex
	session  *concurrency.Session
	override = false
)

// Override overrides the package. Calls to Do will always result in Do's
// argument being executed. Meant for testing purposes.
func Override() {
	pkgMu.Lock()
	defer pkgMu.Unlock()
	override = true
}

// Initialize intializes the package, triggering an initial election.
func Initialize(c *clientv3.Client) error {
	pkgMu.Lock()
	defer pkgMu.Unlock()
	override = false
	var err error
	session, err = concurrency.NewSession(c)
	if err != nil {
		return err
	}
	initSupervisor(session)
	return nil
}

func initSupervisor(session *concurrency.Session) {
	super = newSupervisor(session)
	super.Start()
}

// Resign resigns from leadership if the node holds it. To be used on shutdown.
// Once Resign is called, all further calls to Do will result in an error.
func Resign() error {
	pkgMu.Lock()
	defer pkgMu.Unlock()
	var err error
	if super != nil {
		err = super.Stop()
	}
	super = nil
	if err != nil {
		return err
	}
	return session.Close()
}
