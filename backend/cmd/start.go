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
)

func init() {
	rootCmd.AddCommand(newStartCommand())
}

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			sensuBackend, err := backend.NewBackend(&backend.Config{
				Port:     listenPort,
				StateDir: stateDir,
			})
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

	return cmd
}
