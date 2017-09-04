package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ImportCommand adds a command that allows user to create new handlers from STDIN
func ImportCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "import",
		Short:        "create new handlers from STDIN",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			stat, _ := cli.InFile.Stat()

			// If no data is present in STDIN print out usage
			if stat.Mode()&os.ModeNamedPipe == 0 {
				cmd.Help()
				return nil
			}

			handler := &types.Handler{}
			dec := json.NewDecoder(bufio.NewReader(cli.InFile))
			dec.Decode(handler)

			if err := handler.Validate(); err != nil {
				return err
			}

			err := cli.Client.CreateHandler(handler)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Imported")
			return nil
		},
	}
}
