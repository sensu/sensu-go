package check

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ImportCommand adds command that allows user to create new checks from stdin
func ImportCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:          "import",
		Short:        "create new checks from STDIN",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			stat, _ := os.Stdin.Stat()
			if stat.Mode()&os.ModeNamedPipe == 0 {
				cmd.Help() // Print out usage
				return nil
			}

			check := &types.Check{}
			dec := json.NewDecoder(bufio.NewReader(os.Stdin))
			dec.Decode(check)

			if err := check.Validate(); err != nil {
				return err
			}

			err := cli.Client.CreateCheck(check)
			if err != nil {
				return err
			}

			return nil
		},
	}
}
