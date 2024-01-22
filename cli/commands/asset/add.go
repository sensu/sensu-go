package asset

import (
	"bytes"
	"errors"
	"fmt"
	"path"

	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"

	goversion "github.com/hashicorp/go-version"
	"github.com/sensu/sensu-go/cli/resource"
	"github.com/sensu/sensu-go/types/compat"
)

var rename string
var help string = `
You have successfully added the Sensu asset resource, but the asset will not get downloaded until
it's invoked by another Sensu resource (ex. check). To add this runtime asset to the appropriate
resource, populate the "runtime_assets" field with`

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
		//sudhanshu/5009
		//check if asset.db exits if not then re-create it also delete the SHA associated with it and re-create it.
		//
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

		resources, err := resource.Parse(bytes.NewReader([]byte(asset)))
		if err != nil {
			return err
		}
		if err := resource.Validate(resources, cli.Config.Namespace()); err != nil {
			return err
		}
		assetPath := path.Join(bAsset.Namespace, bAsset.Name)
		for i := range resources {
			meta := compat.GetObjectMeta(resources[i].Value)
			if rename != "" {
				meta.Name = rename
			} else {
				meta.Name = assetPath
			}
			compat.SetObjectMeta(resources[i].Value, meta)
		}
		processor := resource.NewPutter()
		if err := processor.Process(cli.Client, resources); err != nil {
			return err
		}

		fmt.Printf("added asset: %s:%s\n", assetPath, bonsaiVersion.Original())
		assetName := rename
		if assetName == "" {
			assetName = assetPath
		}
		fmt.Printf("%s [\"%s\"].\n", help, assetName)
		return nil
	}
}
