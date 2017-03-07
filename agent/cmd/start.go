package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/sensu/sensu-go/agent"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newStartCommand())
}

var (
	backendURL string
)

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			sensuAgent := agent.NewAgent(&agent.Config{
				BackendURL: backendURL,
			})
			if err := sensuAgent.Run(); err != nil {
				return err
			}

			sigs := make(chan os.Signal, 1)
			done := make(chan struct{}, 1)

			signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
			go func() {
				sig := <-sigs
				log.Println("signal received: ", sig)
				sensuAgent.Stop()
				done <- struct{}{}
			}()

			<-done
			return nil
		},
	}

	cmd.Flags().StringVarP(&backendURL, "backend-url", "b", "ws://localhost:8080", "ws/wss URL of Sensu backend server(s)")

	return cmd
}
