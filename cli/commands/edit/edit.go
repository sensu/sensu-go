package edit

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strings"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/cli/resource"
	"github.com/sensu/core/v3/types"
	"github.com/sensu/sensu-go/util/compat"
	"github.com/spf13/cobra"
)

const (
	// vi is the default editor!
	defaultEditor = "vi"
	// except on Windows
	defaultWindowsEditor = "notepad.exe"
)

func extension(format string) string {
	switch format {
	case config.FormatJSON:
		return "json"
	default:
		return "yaml"
	}
}

type lifter interface {
	Lift() corev3.Resource
}

type namespaceFormat interface {
	Namespace() string
	Format() string
}

type client interface {
	Get(string, interface{}) error
}

func dumpResource(client client, cfg namespaceFormat, typeName string, key []string, to io.Writer) error {
	// Determine the requested resource type. We will use this resource only to
	// determine it's path in the store
	requested, err := resource.Resolve(typeName)
	if err != nil {
		return fmt.Errorf("invalid resource type: %s", typeName)
	}

	switch r := requested.(type) {
	case *corev2.Event:
		// Need an exception for event, because it's a special little type
		if len(key) != 2 {
			return errors.New("events need an entity and check component")
		}
		r.Entity = &corev2.Entity{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: cfg.Namespace(),
				Name:      key[0],
			},
		}
		r.Check = &corev2.Check{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: cfg.Namespace(),
				Name:      key[1],
			},
		}
	case *corev2.Check:
		// Special case here takes care of the check naming boondoggle
		requested = &corev2.CheckConfig{}
		if len(key) != 1 {
			return errors.New("resource name missing")
		}
		requested.SetMetadata(&corev2.ObjectMeta{
			Namespace: cfg.Namespace(),
			Name:      key[0],
		})
	default:
		if len(key) != 1 {
			return errors.New("resource name missing")
		}
		requested.SetMetadata(&corev2.ObjectMeta{
			Namespace: cfg.Namespace(),
			Name:      key[0],
		})
	}
	if lifter, ok := requested.(lifter); ok {
		requested = lifter.Lift()
	}

	response := &types.Wrapper{}

	if err := client.Get(requested.URIPath(), response); err != nil {
		return err
	}

	// Retrieve the concrete resource value from the response
	resource := response.Value.(corev3.Resource)

	format := cfg.Format()
	switch format {
	case "json":
		return helpers.PrintResourceJSON(resource, to)
	default:
		return helpers.PrintYAML([]corev3.Resource{resource}, to)
	}
}

func dumpBlank(cfg namespaceFormat, typeName string, to io.Writer) error {
	resource, err := resource.Resolve(typeName)
	if err != nil {
		return fmt.Errorf("invalid resource type: %s", typeName)
	}
	switch r := resource.(type) {
	case *corev2.Event:
		r.Entity = &corev2.Entity{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: cfg.Namespace(),
			},
		}
		r.Check = &corev2.Check{
			ObjectMeta: corev2.ObjectMeta{
				Namespace: cfg.Namespace(),
			},
		}
	case *corev2.Check:
		// Special case here takes care of the check naming boondoggle
		resource = &corev2.CheckConfig{}
		resource.SetMetadata(&corev2.ObjectMeta{
			Namespace: cfg.Namespace(),
		})
	default:
		resource.SetMetadata(&corev2.ObjectMeta{
			Namespace: cfg.Namespace(),
		})
	}
	if lifter, ok := resource.(lifter); ok {
		resource = lifter.Lift()
	}
	format := cfg.Format()
	switch format {
	case "json":
		return helpers.PrintResourceJSON(resource, to)
	default:
		return helpers.PrintYAML([]corev3.Resource{resource}, to)
	}
}

func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [RESOURCE TYPE] [KEY]...",
		Short: "Edit resources interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			blank, err := cmd.Flags().GetBool("blank")
			if err != nil {
				return err
			}
			if len(args) < 2 && !blank {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			} else if len(args) < 1 && blank {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			tf, err := ioutil.TempFile("", fmt.Sprintf("sensu-resource.*.%s", extension(cli.Config.Format())))
			if err != nil {
				return err
			}
			defer os.Remove(tf.Name())
			orig := new(bytes.Buffer)
			writer := io.MultiWriter(orig, tf)
			if blank {
				if err := dumpBlank(cli.Config, args[0], writer); err != nil {
					return err
				}
			} else {
				if err := dumpResource(cli.Client, cli.Config, args[0], args[1:], writer); err != nil {
					return err
				}
			}
			if err := tf.Close(); err != nil {
				return err
			}
			editorEnv := os.Getenv("EDITOR")
			if strings.TrimSpace(editorEnv) == "" {
				editorEnv = defaultEditor
				if runtime.GOOS == "windows" {
					editorEnv = defaultWindowsEditor
				}
			}
			editorArgs := parseCommand(editorEnv)
			execCmd := exec.Command(editorArgs[0], append(editorArgs[1:], tf.Name())...)
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			if err := execCmd.Run(); err != nil {
				return err
			}
			changedBytes, err := ioutil.ReadFile(tf.Name())
			if err != nil {
				return err
			}
			if bytes.Equal(orig.Bytes(), changedBytes) {
				return nil
			}
			resources, err := resource.Parse(bytes.NewReader(changedBytes))
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				return errors.New("no resources were parsed")
			}
			if err := resource.Validate(resources, cli.Config.Namespace()); err != nil {
				return err
			}
			processor := resource.NewPutter()
			if err := processor.Process(cli.Client, resources); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", compat.URIPath(resources[0].Value))
			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())
	_ = cmd.Flags().BoolP("blank", "b", false, "edit a blank resource, and create it on save")

	return cmd
}

func parseCommand(cmd string) []string {
	scanner := bufio.NewScanner(strings.NewReader(cmd))
	scanner.Split(bufio.ScanWords)
	var result []string
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		// unlikely
		panic(err)
	}
	return result
}
