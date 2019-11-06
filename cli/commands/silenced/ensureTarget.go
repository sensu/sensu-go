package silenced

import (
	"errors"

	"github.com/spf13/cobra"
)

func ensureTarget(cmd *cobra.Command, args []string) error {
	if len(args) > 0 {
		return nil
	}
	flags := cmd.Flags()
	sub, err := flags.GetString("subscription")
	if err != nil {
		return err
	}
	check, err := flags.GetString("check")
	if err != nil {
		return err
	}
	if sub == "" && check == "" {
		target, err := getTarget("specify subscription, check, or both")
		if err != nil {
			return err
		}
		if target == nil || target.Subscription == "" || target.Check == "" {
			return errors.New("Must specificy subscription, check, or both")
		}
	}
	return nil
}
