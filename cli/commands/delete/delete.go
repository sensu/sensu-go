package delete

import (
	"errors"
	"fmt"
	"io"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/resource"
	"github.com/sensu/core/v3/types"
	"github.com/sensu/sensu-go/util/compat"
	"github.com/spf13/cobra"
)

// DeleteCommand deletes generic Sensu resources.
func DeleteCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [-f FILE]",
		Short: "Delete resources from file or STDIN",
		RunE:  execute(cli),
	}

	_ = cmd.Flags().StringP("file", "f", "", "File to delete resources from")

	return cmd
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var in io.Reader
		if len(args) > 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		fp, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}

		in, err = helpers.InputData(fp)
		if err != nil {
			return err
		}

		resources, err := resource.Parse(in)
		if err != nil {
			return err
		}
		if err := resource.Validate(resources, cli.Config.Namespace()); err != nil {
			return err
		}

		return DeleteResources(cli.Client, resources)
	}
}

// DeleteResources deletes all of the parsed resources.
func DeleteResources(client client.GenericClient, resources []*types.Wrapper) error {
	for i, resource := range resources {
		path := compat.URIPath(resource.Value)
		if err := client.Delete(path); err != nil {
			return fmt.Errorf("error deleting resource %d (%s): %s", i, path, err)
		}
	}
	return nil
}
