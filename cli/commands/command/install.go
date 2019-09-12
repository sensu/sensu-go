package command

import (
	"errors"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/cmdmanager"
	"github.com/spf13/cobra"
)

var assetURL string
var assetChecksum string

// InstallCommand adds command that allows a user to install a command plugin
// via Bonsai or a URL.
func InstallCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install [ALIAS] [NAME][:VERSION]",
		Short: "installs a sensuctl command from Bonsai or a URL",
		RunE:  installCommandExecute(cli),
	}

	cmd.Flags().StringVarP(&assetURL, "url", "", "", "specifies a URL to fetch a sensuctl command asset from")
	cmd.Flags().StringVarP(&assetChecksum, "checksum", "", "", "specifies the checksum (SHA512) of a URL asset")

	return cmd
}

func installCommandExecute(cli *cli.SensuCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// If no name is present print out usage
		if len(args) < 1 || len(args) > 2 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		if len(args) == 1 {
			if assetURL == "" {
				return errors.New("must specify a Bonsai asset name or use --url flag")
			}
			if assetChecksum == "" {
				return errors.New("must specify checksum (SHA512) with --checksum when using --url")
			}
		}

		alias := args[0]

		manager, err := cmdmanager.NewCommandManager()
		if err != nil {
			return err
		}

		if len(args) == 2 {
			if err = manager.InstallCommandFromBonsai(alias, args[1]); err != nil {
				return err
			}
		} else {
			if err = manager.InstallCommandFromURL(alias, assetURL, assetChecksum); err != nil {
				return err
			}
		}

		return nil
	}
}
