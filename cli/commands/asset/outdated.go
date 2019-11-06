package asset

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/elements/table"
	"github.com/sensu/sensu-go/types"
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

		bonsaiClient := bonsai.New(bonsai.Config{})

		outdatedAssets := []bonsai.OutdatedAsset{}

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
					outdatedAssets = append(outdatedAssets, bonsai.OutdatedAsset{
						Name:           bonsaiName,
						Namespace:      bonsaiNamespace,
						AssetName:      asset.Name,
						CurrentVersion: installedVersion.String(),
						LatestVersion:  latestVersion.String(),
					})
				}
			}
		}

		// Print the results based on user preferences
		resources := []corev2.Resource{}
		for _, outdatedAsset := range outdatedAssets {
			resources = append(resources, &outdatedAsset)
		}

		return helpers.PrintList(cmd, cli.Config.Format(), printOutdatedToTable, resources, outdatedAssets, header)
	}
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
				return fmt.Sprintf("%s/%s", outdatedAsset.Namespace, outdatedAsset.Name)
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
