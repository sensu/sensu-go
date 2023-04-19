package asset

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/cli"
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
			askOpts := []survey.AskOpt{
				survey.WithStdio(cli.InFile, cli.OutFile, cli.ErrFile),
			}
			if err := opts.administerQuestionnaire(askOpts...); err != nil {
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

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Updated")
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

func (opts *assetOptions) copyFrom(a *v2.Asset) {
	opts.URL = a.URL
	opts.Sha512 = a.Sha512
}

func (opts *assetOptions) copyTo(a *v2.Asset) {
	a.URL = opts.URL
	a.Sha512 = opts.Sha512
}

func (opts *assetOptions) administerQuestionnaire(askOpts ...survey.AskOpt) error {
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

	return survey.Ask(qs, opts, askOpts...)
}
