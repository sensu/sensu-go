package create

import (
	"errors"
	"net/http"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/resource"
	"github.com/spf13/cobra"
)

// CreateCommand creates generic Sensu resources.
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [-r] [[-f URL] ... ]",
		Short: "Create or replace resources from file or URL (path, file://, http[s]://), or STDIN otherwise.",
		RunE:  execute(cli),
	}

	_ = cmd.Flags().StringSliceP("file", "f", nil, "Files, directories, or URLs to create resources from")
	_ = cmd.Flags().BoolP("recursive", "r", false, "Follow subdirectories")

	return cmd
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}
		t := &http.Transport{}
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		client := &http.Client{Transport: t}
		inputs, err := cmd.Flags().GetStringSlice("file")
		if err != nil {
			return err
		}
		processor := resource.NewManagedByLabelPutter("sensuctl")
		if len(inputs) == 0 {
			return resource.ProcessStdin(cli, client, processor)
		}
		recurse, err := cmd.Flags().GetBool("recursive")
		if err != nil {
			return err
		}
		if err := resource.Process(cli, client, inputs, recurse, processor); err != nil {
			return err
		}
		return nil
	}
}
