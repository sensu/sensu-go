package command

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/cmdmanager"
	"github.com/spf13/cobra"
)

// ExecCommand provides a method to execute command plugins.
func ExecCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exec [ALIAS] [args]",
		Short: "executes a sensuctl command plugin",
		RunE:  execCommandExecute(cli),
	}

	return cmd
}

func execCommandExecute(cli *cli.SensuCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// If no name is present print out usage
		if len(args) < 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		manager, err := cmdmanager.NewCommandManager(cli)
		if err != nil {
			return err
		}

		// refresh the access token
		tokens, err := cli.Client.RefreshAccessToken(cli.Config.Tokens())
		if err != nil {
			return err
		}

		// save new tokens to disk
		if err := cli.Config.SaveTokens(tokens); err != nil {
			return err
		}

		commandEnv := []string{
			fmt.Sprintf("SENSU_API_URL=%s", cli.Config.APIUrl()),
			fmt.Sprintf("SENSU_NAMESPACE=%s", cli.Config.Namespace()),
			fmt.Sprintf("SENSU_FORMAT=%s", cli.Config.Format()),
			fmt.Sprintf("SENSU_API_KEY=%s", cli.Config.APIKey()),
			fmt.Sprintf("SENSU_ACCESS_TOKEN=%s", cli.Config.Tokens().Access),
			fmt.Sprintf("SENSU_ACCESS_TOKEN_EXPIRES_AT=%d", cli.Config.Tokens().ExpiresAt),
			fmt.Sprintf("SENSU_REFRESH_TOKEN=%s", cli.Config.Tokens().Refresh),
			fmt.Sprintf("SENSU_TRUSTED_CA_FILE=%s", cli.Config.TrustedCAFile()),
			fmt.Sprintf("SENSU_INSECURE_SKIP_TLS_VERIFY=%s", strconv.FormatBool(cli.Config.InsecureSkipTLSVerify())),
			fmt.Sprintf("SENSU_TIMEOUT=%s", cli.Config.Timeout().String()),
		}

		ctx := context.TODO()
		if err = manager.ExecCommand(ctx, args[0], args[1:], commandEnv); err != nil {
			return err
		}

		return nil
	}
}
