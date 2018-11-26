package edit

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/create"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// vi is the default editor!
const defaultEditor = "vi"

func extension(format string) string {
	switch format {
	case config.FormatYAML:
		return "yaml"
	case config.FormatJSON, config.FormatWrappedJSON:
		return "json"
	default:
		return "txt"
	}
}

func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [RESOURCE TYPE] [KEY]...",
		Short: "Edit resources interactively",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				_ = cmd.Help()
				return errors.New("invalid argument(s) received")
			}
			resourceType := args[0]
			resourceKeyParams := args[1:]
			format := cli.Config.Format()
			switch format {
			case "yaml", "wrapped-json":
			default:
				format = "yaml"
			}
			ctlArgs := []string{
				resourceType,
				"info",
			}
			ctlArgs = append(ctlArgs, resourceKeyParams...)
			ctlArgs = append(ctlArgs, "--format")
			ctlArgs = append(ctlArgs, format)
			originalBytes, err := exec.Command(os.Args[0], ctlArgs...).CombinedOutput()
			if err != nil {
				_, _ = cmd.OutOrStdout().Write(originalBytes)
				return fmt.Errorf("couldn't get resource: %s", err)
			}
			tf, err := ioutil.TempFile("", fmt.Sprintf("sensu-resource.*.%s", extension(format)))
			if err != nil {
				return err
			}
			if _, err := tf.Write(originalBytes); err != nil {
				return err
			}
			if err := tf.Close(); err != nil {
				return err
			}
			editorEnv := os.Getenv("EDITOR")
			if strings.TrimSpace(editorEnv) == "" {
				editorEnv = defaultEditor
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
			if bytes.Equal(originalBytes, changedBytes) {
				return nil
			}
			resources, err := create.ParseResources(bytes.NewReader(changedBytes))
			if err != nil {
				return err
			}
			if len(resources) == 0 {
				return errors.New("no resources were parsed")
			}
			if err := create.ValidateResources(resources); err != nil {
				return err
			}
			if err := create.PutResources(cli.Client, resources); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s\n", resources[0].URIPath())
			return nil
		},
	}

	helpers.AddFormatFlag(cmd.Flags())

	return cmd
}

func parseCommand(cmd string) []string {
	parts := strings.Split(cmd, " ")
	result := []string{}
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) > 0 {
			result = append(result, part)
		}
	}
	return result
}
