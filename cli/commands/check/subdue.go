package check

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

var kitchenTZRE = regexp.MustCompile(`[0-1]?[0-9]:[0-5][0-9]\s?(AM|PM)( .+)?`)

// SubdueCommand adds a command that allows a user to subdue a check
func SubdueCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "subdue NAME",
		Short:        "subdue checks from file or stdin",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Print usage if we do not receive one argument
			if len(args) != 1 {
				return cmd.Help()
			}

			check, err := cli.Client.FetchCheck(args[0])
			if err != nil {
				return err
			}

			subduePath, _ := cmd.Flags().GetString("file")
			var in *os.File

			if len(subduePath) > 0 {
				in, err = os.Open(subduePath)
				if err != nil {
					return err
				}
				
				defer func() { _ = in.Close() }()
			} else {
				in = os.Stdin
			}
			var timeWindows types.TimeWindowWhen
			if err := json.NewDecoder(in).Decode(&timeWindows); err != nil {
				return err
			}
			for _, windows := range timeWindows.MapTimeWindows() {
				for _, window := range windows {
					if err := convertToUTC(window); err != nil {
						return err
					}
				}
			}
			check.Subdue = &timeWindows
			if err := check.Validate(); err != nil {
				return err
			}
			if err := cli.Client.UpdateCheck(check); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "OK")
			return nil
		},
	}

	cmd.Flags().StringP("file", "f", "", "Subdue definition file")

	return cmd
}

func convertToUTC(t *types.TimeWindowTimeRange) error {
	begin, err := offsetTime(t.Begin)
	if err != nil {
		return nil
	}
	end, err := offsetTime(t.End)
	if err != nil {
		return nil
	}
	t.Begin = begin
	t.End = end
	return nil
}

func offsetTime(s string) (string, error) {
	ts, tz, err := extractLocation(s)
	if err != nil {
		return "", err
	}
	tm, err := time.ParseInLocation(time.Kitchen, ts, tz)
	if err != nil {
		return "", err
	}
	_, offset := tm.Zone()
	tm = tm.Add(-time.Duration(offset) * time.Second)
	return tm.Format(time.Kitchen), nil
}

func extractLocation(s string) (string, *time.Location, error) {
	tz := time.Local
	beginMatches := kitchenTZRE.FindStringSubmatch(s)
	possibleTZ := strings.TrimSpace(beginMatches[len(beginMatches)-1])
	if len(possibleTZ) == 0 {
		return s, tz, nil
	}
	loc, err := time.LoadLocation(possibleTZ)
	trimmed := strings.TrimSpace(strings.TrimSuffix(s, possibleTZ))
	normalized := strings.Replace(strings.Replace(trimmed, " AM", "AM", 0), " PM", "PM", 0)
	return normalized, loc, err
}
