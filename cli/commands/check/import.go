package check

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

// ImportCommand adds command that allows user to create new checks from stdin
func ImportCommand(cli *cli.SensuCli) *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "create new checks from STDIN",
		RunE: func(cmd *cobra.Command, args []string) error {
			stat, _ := os.Stdin.Stat()
			if stat.Mode()&os.ModeNamedPipe == 0 {
				cmd.HelpFunc()(cmd, args)
				return nil
			}

			check := &types.Check{}
			dec := json.NewDecoder(bufio.NewReader(os.Stdin))
			dec.Decode(check)

			if err := check.Validate(); err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			err := cli.Client.CreateCheck(check)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}

			return nil
		},
	}
}
