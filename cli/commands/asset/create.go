package asset

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateCommand adds command that allows user to create new assets
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	exec := createExecutor{client: cli.Client}
	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "create new assets",
		RunE:  exec.run,
	}

	cmd.Flags().StringP("url", "u", "", "the URL of the asset")
	cmd.Flags().StringSliceP("metadata", "m", []string{}, "metadata associated with asset")

	return cmd
}

type createExecutor struct {
	client client.APIClient
}

func (e *createExecutor) run(cmd *cobra.Command, args []string) error {
	asset, err := configureNewAsset(cmd.Flags(), args)
	if err != nil {
		return err
	}

	if err := asset.Validate(); err != nil {
		return err
	}

	if err := e.client.CreateAsset(asset); err != nil {
		return err
	}

	fmt.Fprintf(cmd.OutOrStdout(), "OK")
	return nil
}

func configureNewAsset(flags *pflag.FlagSet, args []string) (*types.Asset, error) {
	opts := assetOpts{asset: &types.Asset{}}
	opts.configure(flags, args)
	return opts.asset, opts.err
}

type assetOpts struct {
	asset *types.Asset
	err   error
}

func (o *assetOpts) configure(flags *pflag.FlagSet, args []string) {
	o.setName(args)
	o.setURL(flags)
	o.setMeta(flags)
}

func (e *assetOpts) setName(args []string) {
	if len(args) == 1 {
		e.asset.Name = args[0]
	} else if len(args) > 1 {
		e.err = errors.New("too many arguments given.")
	} else {
		e.err = errors.New("please provide a name for given asset.")
	}
}

func (e *assetOpts) setURL(flags *pflag.FlagSet) {
	if url, err := flags.GetString("url"); err != nil {
		e.err = err
	} else {
		e.asset.URL = url
	}
}

func (e *assetOpts) setMeta(flags *pflag.FlagSet) {
	metadata, err := flags.GetStringSlice("metadata")
	if err != nil {
		e.err = err
	}

	e.asset.Metadata = make(map[string]string, len(metadata))
	for _, meta := range metadata {
		// TODO(james): naive
		splitMeta := strings.SplitAfterN(meta, ":", 2)

		if len(splitMeta) == 2 {
			key := strings.TrimSpace(strings.TrimRight(splitMeta[0], ":"))
			val := strings.TrimSpace(splitMeta[1])
			e.asset.Metadata[key] = val
		} else {
			err := fmt.Sprintf(
				"Metadata value '%s' appears invalid;"+
					"should be in format 'KEY: VALUE'.",
				splitMeta,
			)
			e.err = errors.New(err)
			break
		}
	}
}
