package edit

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/commands/create"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/spf13/cobra"
)

// vi is the default editor!
const defaultEditor = "vi"

type unmarshalFunc func([]byte, interface{}) error

func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit [RESOURCE NAME] [KEY1] ... [KEYN]",
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
			fmt.Println(ctlArgs)
			b, err := exec.Command("sensuctl", ctlArgs...).CombinedOutput()
			if err != nil {
				cmd.OutOrStdout().Write(b)
				return fmt.Errorf("couldn't get resource: %s", err)
			}
			tf, err := ioutil.TempFile("", "sensu-resource-*")
			if err != nil {
				return err
			}
			if _, err := tf.Write(b); err != nil {
				return err
			}
			if err := tf.Close(); err != nil {
				return err
			}
			editorEnv := os.Getenv("EDITOR")
			if editorEnv == "" {
				editorEnv = defaultEditor
			}
			execCmd := exec.Command(editorEnv, tf.Name())
			execCmd.Stdin = os.Stdin
			execCmd.Stdout = os.Stdout
			execCmd.Stderr = os.Stderr
			if err := execCmd.Run(); err != nil {
				return err
			}
			tf, err = os.Open(tf.Name())
			if err != nil {
				return err
			}
			resources, err := create.ParseResources(tf)
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
