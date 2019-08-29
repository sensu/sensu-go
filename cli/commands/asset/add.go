package asset

import (
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// AddCommand adds c ommand that allows user to add assets from Bonsai.
func AddCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [NAME]",
		Short: "adds an asset definition fetched from Bonsai",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no name is present print out usage
			if len(args) != 1 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}

			name := args[0]

			bAsset, err := corev2.NewBonsaiBaseAsset(name)
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

			bonsaiClient := bonsai.New(bonsai.BonsaiConfig{})
			asset, err := bonsaiClient.FetchAsset(bAsset.Namespace, bAsset.Name)
			if err != nil {
				return err
			}

			if version == nil {
				fmt.Println("no version specified, using latest:", asset.LatestVersion())
				version = asset.LatestVersion()
			} else if asset.HasVersion(version) {
				fmt.Println("a release for the specified version exists!")
			} else {
				return fmt.Errorf("version \"%s\" of asset \"%s/%s\" does not exist", version, bAsset.Namespace, bAsset.Name)
			}

			fmt.Printf("namespace: %s, name: %s, version: %s\n", bAsset.Namespace, bAsset.Name, version)

			return nil
		},
	}

	return cmd
}
