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
)

func init() {
	rootCmd.AddCommand(newStartCommand())
}

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu backend",
		RunE: func(cmd *cobra.Command, args []string) error {
			sensuBackend := backend.NewBackend(&backend.Config{
				Port: listenPort,
			})
			if err := sensuBackend.Run(); err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)
			done := make(chan struct{}, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				log.Println("signal received: ", sig)
				sensuBackend.Stop()
				done <- struct{}{}
			}()

			<-done
			return nil
		},
	}

	cmd.Flags().IntVarP(&listenPort, "port", "p", 8080, "port to listen on")

	return cmd
}
