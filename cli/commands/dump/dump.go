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
	"github.com/spf13/cobra"
)

const (
	// VerbList is the sensuctl verb to list resources.
	VerbList = "list"
	// VerbInfo is the sensuctl verb to get info about a resource.
	VerbInfo = "info"
)

// Action is a resource/verb tuple for sensuctl commands.
type Action struct {
	Resource   string
	Verb       string
	Namespaced bool
}

var (
	// All is all the core resource types and associated sensuctl verbs (non-namespaced resources are intentionally ordered first).
	All = []Action{
		Action{Resource: "namespace", Verb: VerbList, Namespaced: false},
		Action{Resource: "cluster-role", Verb: VerbList, Namespaced: false},
		Action{Resource: "cluster-role-binding", Verb: VerbList, Namespaced: false},
		Action{Resource: "user", Verb: VerbList, Namespaced: false},
		Action{Resource: "tessen", Verb: VerbInfo, Namespaced: false},
		Action{Resource: "asset", Verb: VerbList, Namespaced: true},
		Action{Resource: "check", Verb: VerbList, Namespaced: true},
		Action{Resource: "entity", Verb: VerbList, Namespaced: true},
		Action{Resource: "event", Verb: VerbList, Namespaced: true},
		Action{Resource: "filter", Verb: VerbList, Namespaced: true},
		Action{Resource: "handler", Verb: VerbList, Namespaced: true},
		Action{Resource: "hook", Verb: VerbList, Namespaced: true},
		Action{Resource: "mutator", Verb: VerbList, Namespaced: true},
		Action{Resource: "role", Verb: VerbList, Namespaced: true},
		Action{Resource: "role-binding", Verb: VerbList, Namespaced: true},
		Action{Resource: "silenced", Verb: VerbList, Namespaced: true},
	}
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

		// parse the comma separated resource types and match against the defined actions
		var actions []Action
		if args[0] == "all" {
			actions = All
		} else {
			types := strings.Split(args[0], ",")
			for _, t := range types {
				for _, action := range All {
					if t == action.Resource {
						actions = append(actions, action)
					}
				}
			}
		}

		// iterate the matched actions and start building a sensuctl command
		var out string
		for _, a := range actions {
			ctlArgs := []string{
				a.Resource,
				a.Verb,
				"--format",
				format,
			}

			// append --namespace or --all-namespaces flag if compatible with the resource type
			if a.Namespaced {
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
			if format == "wrapped-json" {
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
