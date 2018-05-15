package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/pflag"

	"github.com/atlassian/gostatsd/pkg/cluster/nodes"
)

// Cluster is everything for running a single node in a cluster
type Cluster struct {
	RedisAddr      string
	Namespace      string
	Target         string
	UpdateInterval time.Duration
	ExpiryInterval time.Duration
}

// newCluster will create a new Cluster with default values.
func newCluster() *Cluster {
	local, err := nodes.LocalAddress("1.1.1.1:1")

	if err != nil {
		return nil
	}

	return &Cluster{
		RedisAddr:      "127.0.0.1:6379",
		Namespace:      "namespace",
		Target:         local.String() + ":8125",
		UpdateInterval: time.Second,
		ExpiryInterval: 30 * time.Second,
	}
}

// AddFlags adds flags for a specific Server to the specified FlagSet.
func (c *Cluster) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&c.RedisAddr, "redis-addr", c.RedisAddr, "Redis address")
	fs.StringVar(&c.Namespace, "namespace", c.Namespace, "Namespace")
	fs.StringVar(&c.Target, "target", c.Target, "Target port")
	fs.DurationVar(&c.UpdateInterval, "update-interval", c.UpdateInterval, "Cluster update interval")
	fs.DurationVar(&c.ExpiryInterval, "expiry-interval", c.ExpiryInterval, "Cluster expiry interval")
}

// Run runs the specified Cluster.
func (c *Cluster) Run() error {
	rnt := nodes.NewRedisNodeTracker(c.RedisAddr, c.Namespace, c.Target, c.UpdateInterval, c.ExpiryInterval)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go rnt.Run(ctx)

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for range t.C {
		nodes := rnt.List()
		fmt.Printf("%v\n", nodes)
	}

	return nil
}
