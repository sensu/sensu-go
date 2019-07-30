package dump

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	utilstrings "github.com/sensu/sensu-go/util/strings"
	"github.com/spf13/cobra"
)

var (
	// All is all the core resource types that sensuctl can list (non-namespaced resources are intentionally ordered first).
	All = []string{"namespace", "cluster-role", "cluster-role-binding", "user", "asset", "check", "entity", "event", "filter", "handler", "hook", "mutator", "role", "role-binding", "silenced"}
	// NoNamespace is all the non-namespaced core resource types that sensuctl can list.
	NoNamespace = []string{"namespace", "cluster-role", "cluster-role-binding", "user"}
)

// Command dumps generic Sensu resources to a file or STDOUT.
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dump [RESOURCE TYPE],[RESOURCE TYPE]... [-f FILE]",
		Short: "dump resources to a file or STDOUT",
		RunE:  execute(cli),
	}

	helpers.AddFormatFlag(cmd.Flags())
	helpers.AddAllNamespace(cmd.Flags())
	_ = cmd.Flags().StringP("file", "f", "", "file to dump resources to")

	return cmd
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			_ = cmd.Help()
			return errors.New("invalid argument(s) received")
		}

		// get the configured format or the flag override
		format := cli.Config.Format()
		if flag := helpers.GetChangedStringValueFlag("format", cmd.Flags()); flag != "" {
			format = flag
		}
		switch format {
		case "yaml", "wrapped-json":
		default:
			format = "yaml"
		}

		// parse the comma separated resource types
		var types []string
		if args[0] == "all" {
			types = All
		} else {
			types = strings.Split(args[0], ",")
		}

		// iterate the desired types and start building a sensuctl list command
		var out string
		for i, t := range types {
			ctlArgs := []string{
				t,
				"list",
				"--format",
				format,
			}

			// append --namespace or --all-namespaces flag if compatible with the resource type
			if !utilstrings.InArray(t, NoNamespace) {
				if ok, err := cmd.Flags().GetBool(flags.AllNamespaces); err != nil {
					return err
				} else if ok {
					ctlArgs = append(ctlArgs, "--all-namespaces")
				} else {
					ctlArgs = append(ctlArgs, "--namespace")
					ctlArgs = append(ctlArgs, cli.Config.Namespace())
				}
			}

			// execute the command and build wrapped-json or yaml lists
			originalBytes, err := exec.Command(os.Args[0], ctlArgs...).CombinedOutput()
			if err != nil {
				_, _ = cmd.OutOrStdout().Write(originalBytes)
				return fmt.Errorf("couldn't get resource: %s", err)
			}
			if len(originalBytes) == 0 {
				continue
			}
			if format == "wrapped-json" || i == len(types)-1 || out == "" {
				out = fmt.Sprintf("%s%s", out, string(originalBytes))
			} else {
				out = fmt.Sprintf("%s---\n%s", out, string(originalBytes))
			}
		}

		// write data to file or STDOUT
		fp, err := cmd.Flags().GetString("file")
		if err != nil {
			return err
		}
		if fp == "" {
			_, err := fmt.Print(out)
			return err
		}
		f, err := os.Create(fp)
		defer f.Close()
		if err != nil {
			return err
		}
		_, err = f.WriteString(out)
		return err
	}
}
