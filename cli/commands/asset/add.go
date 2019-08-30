package asset

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/bonsai"
	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"

	goversion "github.com/hashicorp/go-version"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli/commands/create"
	"github.com/sensu/sensu-go/types"
)

// AddCommand adds c ommand that allows user to add assets from Bonsai.
func AddCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [NAME]",
		Short: "adds an asset definition fetched from Bonsai",
		RunE:  addCommandExecute(cli),
	}

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
		bonsaiAsset, err := bonsaiClient.FetchAsset(bAsset.Namespace, bAsset.Name)
		if err != nil {
			return err
		}

		if version == nil {
			fmt.Println("no version specified, using latest:", bonsaiAsset.LatestVersion())
			version = bonsaiAsset.LatestVersion()
		} else if bonsaiAsset.HasVersion(version) {
			fmt.Printf("version \"%s\" exists!\n", version)
		} else {
			return fmt.Errorf("version \"%s\" of asset \"%s/%s\" does not exist", version, bAsset.Namespace, bAsset.Name)
		}

		fmt.Printf("fetching asset: %s/%s:%s\n", bAsset.Namespace, bAsset.Name, version)

		asset, err := bonsaiClient.FetchAssetVersion(bAsset.Namespace, bAsset.Name, version.String())
		if err != nil {
			return err
		}

		resources, err := ParseResources(asset)
		if err != nil {
			return err
		}
		if err := create.ValidateResources(resources, cli.Config.Namespace()); err != nil {
			return err
		}
		return create.PutResources(cli.Client, resources)
	}
}

// ParseResources is a rather heroic function that will parse any number of valid
// JSON resources. Since it attempts to be intelligent, it likely contains bugs.
//
// TODO(amdprophet): This is a modified copy of the ParseResources function
// from github.com/sensu/sensu-go/cli/commands/create/create.go. We may want to
// place this logic in a common package.
func ParseResources(jsonStr string) ([]types.Wrapper, error) {
	var resources []types.Wrapper
	var err error

	count := 0
	jsonBytes := []byte(jsonStr)
	dec := json.NewDecoder(bytes.NewReader(jsonBytes))
	dec.DisallowUnknownFields()
	errCount := 0
	for dec.More() {
		var w types.Wrapper
		if rerr := dec.Decode(&w); rerr != nil {
			// Write out as many errors as possible before bailing,
			// but cap it at 10.
			err = errors.New("some resources couldn't be parsed")
			if errCount > 10 {
				err = errors.New("too many errors")
				break
			}
			describeError(count, rerr)
			errCount++
		}
		resources = append(resources, w)
		count++
	}

	// TODO(echlebek): remove this
	filterCheckSubdue(resources)

	return resources, err
}

// copied from github.com/sensu/sensu-go/cli/commands/create/create.go for now
func describeError(index int, err error) {
	jsonErr, ok := err.(*json.UnmarshalTypeError)
	if !ok {
		fmt.Fprintf(os.Stderr, "resource %d: %s\n", index, err)
		return
	}
	fmt.Fprintf(os.Stderr, "resource %d: (offset %d): %s\n", index, jsonErr.Offset, err)
}

// filterCheckSubdue nils out any check subdue fields that are supplied.
// TODO(echlebek): this is temporary; remove it after fixing check subdue.
//
// copied from github.com/sensu/sensu-go/cli/commands/create/create.go for now
func filterCheckSubdue(resources []types.Wrapper) {
	for i := range resources {
		switch val := resources[i].Value.(type) {
		case *types.CheckConfig:
			val.Subdue = nil
		case *types.Check:
			val.Subdue = nil
		case *types.EventFilter:
			val.When = nil
		}
	}
}
