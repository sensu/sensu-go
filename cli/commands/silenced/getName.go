package silenced

import (
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

func getName(cmd *cobra.Command, args []string) (string, error) {
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
	name, err := types.SilencedName(sub, check)
	if err != nil {
		name, err = askName("specify subscription, check, or both")
	}
	return name, err
}
