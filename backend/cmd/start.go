package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sensu/sensu-go/backend"
	"github.com/spf13/cobra"
)

var (
	listenPort int
	stateDir   string

	etcdClientListenURL string
	etcdPeerListenURL   string
	etcdInitialCluster  string
)

func init() {
	rootCmd.AddCommand(newStartCommand())
}

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := &backend.Config{
				Port:     listenPort,
				StateDir: stateDir,
			}

			if etcdClientListenURL != "" {
				cfg.EtcdClientListenURL = etcdClientListenURL
			}

			if etcdPeerListenURL != "" {
				cfg.EtcdPeerListenURL = etcdPeerListenURL
			}

			if etcdInitialCluster != "" {
				cfg.EtcdInitialCluster = etcdInitialCluster
			}

			sensuBackend, err := backend.NewBackend(cfg)
			if err != nil {
				return err
			}
			sensuBackend.Run()

			sigs := make(chan os.Signal, 1)
			done := make(chan struct{}, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				log.Println("signal received: ", sig)
				sensuBackend.Stop()
				done <- struct{}{}
			}()

			select {
			case <-done:
				return nil
			case err := <-sensuBackend.Err():
				return err
			}
		},
	}

	cmd.Flags().IntVarP(&listenPort, "port", "p", 8080, "port to listen on")
	cmd.Flags().StringVarP(&stateDir, "state-dir", "d", "/var/lib/sensu", "path to sensu state storage")

	// For now don't set defaults for these. This allows us to control defaults on NewBackend(). We may wish
	// to do something more interesting here as well--e.g. only expose these settings via some kind of compile
	// feature flag so that they're used only for testing, etc. But for now, we make these easily configurable
	// for end-to-end testing. -grep
	cmd.Flags().StringVar(&etcdClientListenURL, "store-client-url", "", "store client listen URL")
	cmd.Flags().StringVar(&etcdPeerListenURL, "store-peer-url", "", "store peer URL")
	cmd.Flags().StringVar(&etcdInitialCluster, "store-initial-cluster", "", "store initial cluster")

	return cmd
}
