package nodes

import (
	"context"
	"net"
	"time"
)

// NodeTracker is an interface for tracking and selecting nodes in a cluster
type NodeTracker interface {
	// Runs the node tracker until the context is closed.
	Run(ctx context.Context)

	// List returns a list of all nodes being tracked.  The list returned will be
	// a copy of the underlying list of nodes.  Intended for admin interfaces,
	// not performance critical code.  Thread safe.
	List() []string

	// Select will use the provided key to pick a node from the list of tracked
	// nodes and return it.  Returns an error if there are no nodes available.
	// Thread safe.
	Select(key uint64) (string, error)
}

type node struct {
	nodeid string
	expiry time.Time
}

type nodeList []*node

func (nl nodeList) Len() int           { return len(nl) }
func (nl nodeList) Swap(i, j int)      { nl[i], nl[j] = nl[j], nl[i] }
func (nl nodeList) Less(i, j int) bool { return nl[i].nodeid < nl[j].nodeid }

// LocalAddress is a helper function to return the local IP address that would
// be used to connect to a specified target.  Useful to get the IP that should
// be advertised externally.
func LocalAddress(target string) (net.IP, error) {
	// Mostly lifted from https://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
	conn, err := net.Dial("udp", target)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
