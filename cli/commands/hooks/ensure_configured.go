package hooks

import (
	"fmt"
	"os"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

const (
	// ConfigurationRequirement used to identify the annotation flag for this handler
	//
	// Usage:
	//
	//	my_cmd := cobra.Command{
	//		Use: "Setup",
	//		Annotations: map[string]string{
	//			ConfigurationRequirement: ConfigurationNotRequired,
	//		}
	//	}
	ConfigurationRequirement = "CREDENTIALS_REQUIREMENT"

	// ConfigurationNotRequired specifies that the command does not require
	// credentials to be configured to complete operations
	ConfigurationNotRequired = "NO"
)

// ConfigurationPresent - unless the given command specifies that configuration
// is not required, func checks that host & access-token have been configured.
func ConfigurationPresent(cmd *cobra.Command, cli *cli.SensuCli) error {
	// If the command was configured to ignore whether or not the CLI has been
	// configured stop execution.
	if cmd.Annotations[ConfigurationRequirement] == ConfigurationNotRequired {
		return nil
	}

	// Check that both a URL and an access token are present
	tokens := cli.Config.Tokens()

	if cli.Config.APIUrl() == "" {
		return fmt.Errorf(
			"No API URL is defined. You can configure an API URL by running \"%s configure\"",
			os.Args[0],
		)
	}

	if tokens == nil || tokens.Access == "" {
		return fmt.Errorf(
			"Unable to locate credentials. You can configure credentials by running \"%s configure\"",
			os.Args[0],
		)
	}

	return nil
}
