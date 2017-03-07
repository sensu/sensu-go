package main

import (
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
			return nil
		},
	}

	cmd.Flags().StringVarP(&backendURL, "backend-url", "b", "ws://localhost:8080", "ws/wss URL of Sensu backend server(s)")

	return cmd
}
