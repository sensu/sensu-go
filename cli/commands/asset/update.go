package asset

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// UpdateCommand adds command that allows user to create new assets
func UpdateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "update [NAME]",
		Short:        "update assets",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print out usage if we do not receive one argument
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			// Fetch assets from API
			name := args[0]
			asset, err := cli.Client.FetchAsset(name)
			if err != nil {
				return err
			}

			// Administer questionnaire
			opts := newAssetOptions()
			opts.copyFrom(asset)
			if err := opts.administerQuestionnaire(); err != nil {
				return err
			}

			// Apply given arguments to asset
			opts.copyTo(asset)

			if err := asset.Validate(); err != nil {
				return err
			}

			if err := cli.Client.UpdateAsset(asset); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	return cmd
}

type assetOptions struct {
	URL    string
	Sha512 string
}

func newAssetOptions() *assetOptions {
	opts := assetOptions{}
	return &opts
}

func (opts *assetOptions) copyFrom(a *types.Asset) {
	opts.URL = a.URL
	opts.Sha512 = a.Sha512
}

func (opts *assetOptions) copyTo(a *types.Asset) {
	a.URL = opts.URL
	a.Sha512 = opts.Sha512
}

func (opts *assetOptions) administerQuestionnaire() error {
	var qs = []*survey.Question{
		{
			Name:     "url",
			Prompt:   &survey.Input{Message: "URL:", Default: opts.URL},
			Validate: survey.Required,
		},
		{
			Name:     "sha512",
			Prompt:   &survey.Input{Message: "SHA-512 Checksum:", Default: opts.Sha512},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, opts)
}
