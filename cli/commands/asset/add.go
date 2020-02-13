package asset

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"

	goversion "github.com/hashicorp/go-version"
	"github.com/sensu/sensu-go/cli/commands/create"
)

var rename string

// AddCommand adds command that allows user to add assets from Bonsai.
func AddCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [NAME]",
		Short: "adds an asset definition fetched from Bonsai",
		RunE:  addCommandExecute(cli),
	}

	cmd.Flags().StringVarP(&rename, "rename", "r", "", "rename the asset to the provided string after fetching it from Bonsai")

	return cmd
}

func addCommandExecute(cli *cli.SensuCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// If no name is present print out usage
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		name := args[0]

		bAsset, err := bonsai.NewBaseAsset(name)
		if err != nil {
			return err
		}

		var version *goversion.Version
		if bAsset.Version != "" {
			version, err = goversion.NewVersion(bAsset.Version)
			if err != nil {
				return err
			}
		}

		bonsaiClient := bonsai.New(bonsai.Config{})
		bonsaiAsset, err := bonsaiClient.FetchAsset(bAsset.Namespace, bAsset.Name)
		if err != nil {
			return err
		}

		bonsaiVersion, err := bonsaiAsset.BonsaiVersion(version)
		if err != nil {
			return err
		}

		if version == nil {
			fmt.Println("no version specified, using latest:", bonsaiVersion.Original())
		}

		fmt.Printf("fetching bonsai asset: %s/%s:%s\n", bAsset.Namespace, bAsset.Name, bonsaiVersion.Original())

		asset, err := bonsaiClient.FetchAssetVersion(bAsset.Namespace, bAsset.Name, bonsaiVersion.Original())
		if err != nil {
			return err
		}

		resources, err := create.ParseResources(bytes.NewReader([]byte(asset)))
		if err != nil {
			return err
		}
		if err := create.ValidateResources(resources, cli.Config.Namespace()); err != nil {
			return err
		}
		for i := range resources {
			meta := resources[i].Value.GetObjectMeta()
			if rename != "" {
				meta.Name = rename
			} else {
				meta.Name = fmt.Sprintf("%s/%s", bAsset.Namespace, bAsset.Name)
			}
			resources[i].Value.SetObjectMeta(meta)
		}
		if err := create.PutResources(cli.Client, resources); err != nil {
			return err
		}

		fmt.Printf("added asset: %s/%s:%s\n", bAsset.Namespace, bAsset.Name, bonsaiVersion.Original())
		return nil
	}
}
