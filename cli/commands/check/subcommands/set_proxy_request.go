package subcommands

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// SetProxyRequestsCommand adds a command that allows a user to set the proxy
// requests for a check
func SetProxyRequestsCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "set-proxy-requests [NAME]",
		Short:        "set proxy requests for a check from file or stdin",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			check, err := cli.Client.FetchCheck(args[0])
			if err != nil {
				return err
			}

			filePath, _ := cmd.Flags().GetString("file")
			var in *os.File

			if len(filePath) > 0 {
				in, err = os.Open(filePath)
				if err != nil {
					return err
				}

				defer func() { _ = in.Close() }()
			} else {
				in = os.Stdin
			}
			var proxyRequest types.ProxyRequests
			if err := json.NewDecoder(in).Decode(&proxyRequest); err != nil {
				return err
			}
			check.ProxyRequests = &proxyRequest
			if err := check.Validate(); err != nil {
				return err
			}
			if err := cli.Client.UpdateCheck(check); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Proxy request definition file")

	return cmd
}
