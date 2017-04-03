package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sensu/sensu-go/agent"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newStartCommand())
}

var (
	backendURL    string
	agentID       string
	subscriptions string
)

func newStartCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "start the sensu agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := agent.NewConfig()
			cfg.BackendURL = backendURL

			if agentID != "" {
				cfg.AgentID = agentID
			}

			if subscriptions != "" {
				// TODO(greg): we prooobably want someeee sort of input validation.
				cfg.Subscriptions = strings.Split(subscriptions, ",")
			}

			sensuAgent := agent.NewAgent(cfg)
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
	cmd.Flags().StringVar(&agentID, "id", "", "agent ID (defaults to hostname)")
	cmd.Flags().StringVar(&subscriptions, "subscriptions", "", "comma-delimited list of agent subscriptions")

	return cmd
}
