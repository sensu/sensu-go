package silenced

import (
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

func getID(cmd *cobra.Command, args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	flags := cmd.Flags()
	sub, err := flags.GetString("subscription")
	if err != nil {
		return "", err
	}
	check, err := flags.GetString("check")
	if err != nil {
		return "", err
	}
	id, err := types.SilencedID(sub, check)
	if err != nil {
		id, err = askID("specify subscription, check, or both")
	}
	return id, err
}
