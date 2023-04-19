package asset

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/spf13/cobra"
)

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
			namespace = corev2.NamespaceTypeAll
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

		bonsaiClient := bonsai.New(bonsai.Config{})

		// Determine which local assets are outdated
		outdatedAssets, err := outdatedAssets(results, bonsaiClient)
		if err != nil {
			return err
		}

		// Print the results based on user preferences
		resources := []corev2.Resource{}
		for _, outdatedAsset := range outdatedAssets {
			resources = append(resources, &outdatedAsset)
		}

		return helpers.PrintList(cmd, cli.Config.Format(), printOutdatedToTable, resources, outdatedAssets, header)
	}
}

// outdatedAssets compares the local Bonsai assets against the latest versions
// on Bonsai and returns a list of assets that can be upgraded
func outdatedAssets(assets []corev2.Asset, client bonsai.Client) ([]bonsai.OutdatedAsset, error) {
	outdatedAssets := []bonsai.OutdatedAsset{}

	for _, asset := range assets {
		annotations := asset.GetObjectMeta().Annotations
		if annotations[bonsai.URLAnnotation] != "" {
			bonsaiVersion := asset.GetObjectMeta().Annotations[bonsai.VersionAnnotation]
			bonsaiNamespace := asset.GetObjectMeta().Annotations[bonsai.NamespaceAnnotation]
			bonsaiName := asset.GetObjectMeta().Annotations[bonsai.NameAnnotation]

			if bonsaiVersion == "" {
				return nil, fmt.Errorf("asset missing %s annotation: %s", bonsai.VersionAnnotation, asset.Name)
			}
			if bonsaiNamespace == "" {
				return nil, fmt.Errorf("asset missing %s annotation: %s", bonsai.NamespaceAnnotation, asset.Name)
			}
			if bonsaiName == "" {
				return nil, fmt.Errorf("asset missing %s annotation: %s", bonsai.NameAnnotation, asset.Name)
			}

			installedVersion, err := goversion.NewVersion(bonsaiVersion)
			if err != nil {
				return nil, fmt.Errorf("could not parse version %q of asset %s: %s", bonsaiVersion, asset.Name, err)
			}

			bonsaiAsset, err := client.FetchAsset(bonsaiNamespace, bonsaiName)
			if err != nil {
				return nil, fmt.Errorf("could not fetch asset %s: %s", asset.Name, err)
			}

			latestVersion := bonsaiAsset.LatestVersion()
			if latestVersion == nil {
				return nil, fmt.Errorf("could not parse the latest version of asset %s", asset.Name)
			}

			if installedVersion.LessThan(latestVersion) {
				outdatedAssets = append(outdatedAssets, bonsai.OutdatedAsset{
					BonsaiName:      bonsaiName,
					BonsaiNamespace: bonsaiNamespace,
					AssetName:       asset.Name,
					CurrentVersion:  installedVersion.Original(),
					LatestVersion:   latestVersion.Original(),
				})
			}
		}
	}

	return outdatedAssets, nil
}

func printOutdatedToTable(results interface{}, writer io.Writer) {
	table := table.New([]*table.Column{
		{
			Title:       "Asset Name",
			ColumnStyle: table.PrimaryTextStyle,
			CellTransformer: func(data interface{}) string {
				outdatedAsset, ok := data.(bonsai.OutdatedAsset)
				if !ok {
					return cli.TypeError
				}
				return outdatedAsset.AssetName
			},
		},
		{
			Title: "Bonsai Asset",
			CellTransformer: func(data interface{}) string {
				outdatedAsset, ok := data.(bonsai.OutdatedAsset)
				if !ok {
					return cli.TypeError
				}
				return fmt.Sprintf("%s/%s", outdatedAsset.BonsaiNamespace, outdatedAsset.BonsaiName)
			},
		},
		{
			Title: "Current Version",
			CellTransformer: func(data interface{}) string {
				outdatedAsset, ok := data.(bonsai.OutdatedAsset)
				if !ok {
					return cli.TypeError
				}
				return outdatedAsset.CurrentVersion
			},
		},
		{
			Title: "Latest Version",
			CellTransformer: func(data interface{}) string {
				outdatedAsset, ok := data.(bonsai.OutdatedAsset)
				if !ok {
					return cli.TypeError
				}
				return outdatedAsset.LatestVersion
			},
		},
	})

	table.Render(writer, results)
}
