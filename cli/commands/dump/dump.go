package dump

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
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

	helpers.AddAllNamespace(cmd.Flags())
	_ = cmd.Flags().StringP("format", "", cli.Config.Format(), fmt.Sprintf(`format of data returned ("%s"|"%s")`, config.FormatWrappedJSON, config.FormatYAML))
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
		case config.FormatYAML, config.FormatWrappedJSON:
		default:
			format = config.FormatYAML
		}

		// parse the comma separated resource types and match against the defined actions
		var actions []Action
		if args[0] == "all" {
			actions = All
		} else {
			// check for duplicates first
			types := strings.Split(args[0], ",")
			for i := 0; i < len(types); i++ {
				for v := 0; v < i; v++ {
					if types[v] == types[i] {
						return fmt.Errorf("duplicate resource type: %s", types[v])
					}
				}
			}
			// build actions for sensuctl
			for _, t := range types {
				length := len(actions)
				for _, action := range All {
					if t == action.Resource {
						actions = append(actions, action)
					}
				}
				if length == len(actions) {
					return fmt.Errorf("invalid resource type: %s", t)
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
			originalBytes, _ := exec.Command(os.Args[0], ctlArgs...).Output()
			if len(originalBytes) == 0 {
				continue
			}
			if format == config.FormatWrappedJSON {
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
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(out)
		return err
	}
}
