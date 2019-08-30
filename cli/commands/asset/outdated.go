package asset

import (
	"errors"
	"fmt"
	"net/http"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

type OutdatedAsset struct {
	Name           string
	Namespace      string
	AssetName      string
	CurrentVersion string
	LatestVersion  string
}

// OutdatedCommand adds a command that allows users to list outdated assets
// that have been added from Bonsai.
func OutdatedCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "lists any assets installed from Bonsai that have newer versions available",
		RunE:  outdatedCommandExecute(cli),
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())
	helpers.AddFieldSelectorFlag(cmd.Flags())
	helpers.AddLabelSelectorFlag(cmd.Flags())
	helpers.AddChunkSizeFlag(cmd.Flags())

	return cmd
}

func outdatedCommandExecute(cli *cli.SensuCli) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		namespace := cli.Config.Namespace()
		if ok, _ := cmd.Flags().GetBool(flags.AllNamespaces); ok {
			namespace = types.NamespaceTypeAll
		}

		opts, err := helpers.ListOptionsFromFlags(cmd.Flags())
		if err != nil {
			return err
		}

		// Fetch assets from API
		var header http.Header
		results := []corev2.Asset{}
		err = cli.Client.List(client.AssetsPath(namespace), &results, &opts, &header)
		if err != nil {
			return err
		}

		bonsaiClient := bonsai.New(bonsai.BonsaiConfig{})

		outdatedAssets := []OutdatedAsset{}

		for _, asset := range results {
			annotations := asset.GetObjectMeta().Annotations
			if annotations["io.sensu.bonsai.api_url"] != "" {
				bonsaiVersion := asset.GetObjectMeta().Annotations["io.sensu.bonsai.version"]
				bonsaiNamespace := asset.GetObjectMeta().Annotations["io.sensu.bonsai.namespace"]
				bonsaiName := asset.GetObjectMeta().Annotations["io.sensu.bonsai.name"]

				if bonsaiVersion == "" {
					return fmt.Errorf("asset missing io.sensu.bonsai.version annotation: %s", asset.Name)
				}
				if bonsaiNamespace == "" {
					return fmt.Errorf("asset missing io.sensu.bonsai.namespace annotation: %s", asset.Name)
				}
				if bonsaiName == "" {
					return fmt.Errorf("asset missing io.sensu.bonsai.name annotation: %s", asset.Name)
				}

				bonsaiAsset, err := bonsaiClient.FetchAsset(bonsaiNamespace, bonsaiName)
				if err != nil {
					return err
				}

				installedVersion, err := goversion.NewVersion(bonsaiVersion)
				if err != nil {
					return err
				}

				latestVersion := bonsaiAsset.LatestVersion()

				if installedVersion.LessThan(latestVersion) {
					outdatedAssets = append(outdatedAssets, OutdatedAsset{
						Name:           bonsaiName,
						Namespace:      bonsaiNamespace,
						AssetName:      asset.Name,
						CurrentVersion: installedVersion.String(),
						LatestVersion:  latestVersion.String(),
					})
				}
			}
		}

		if len(outdatedAssets) == 0 {
			fmt.Println("all bonsai assets are up to date!")
		} else {
			for _, outdatedAsset := range outdatedAssets {
				assetName := outdatedAsset.AssetName
				if assetName != fmt.Sprintf("%s/%s", outdatedAsset.Namespace, outdatedAsset.Name) {
					assetName = fmt.Sprintf("%s (%s/%s)", outdatedAsset.AssetName, outdatedAsset.Namespace, outdatedAsset.Name)
				}
				fmt.Printf("%s has a newer version (current: %s, latest %s)\n", assetName, outdatedAsset.CurrentVersion, outdatedAsset.LatestVersion)
			}
		}

		return nil
	}
}
