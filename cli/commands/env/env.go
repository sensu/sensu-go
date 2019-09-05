package env

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

const (
	envTmpl = `{{ .Prefix }}SENSU_API_URL{{ .Delimiter }}{{ .APIURL }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_NAMESPACE{{ .Delimiter }}{{ .Namespace }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_FORMAT{{ .Delimiter }}{{ .Format }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_ACCESS_TOKEN{{ .Delimiter }}{{ .AccessToken }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_ACCESS_TOKEN_EXPIRES_AT{{ .Delimiter }}{{ .AccessTokenExpiresAt }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_REFRESH_TOKEN{{ .Delimiter }}{{ .RefreshToken }}{{ .LineEnding }}` // +
	//`{{ .Prefix }}SENSU_TRUSTED_CA_FILE{{ .Delimiter }}{{ .TrustedCAFile }}{{ .LineEnding }}` +
	//`{{ .Prefix }}SENSU_INSECURE_SKIP_TLS_VERIFY{{ .Delimiter }}{{ .InsecureSkipTLSVerify }}{{ .LineEnding }}`

	shellFlag = "shell"
)

// Command display the commands to set up the environment used by sensuctl
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "env",
		Short:   "display the commands to set up the environment used by sensuctl",
		PreRunE: refreshAccessToken(cli),
		RunE:    execute(cli),
	}

	_ = cmd.Flags().StringP(shellFlag, "", "",
		fmt.Sprintf(
			`force environment to be configured for a specified shell ("%s"|"%s"|"%s")`,
			"bash", "cmd", "powershell",
		))

	return cmd
}

type shellConfig struct {
	Prefix     string
	Delimiter  string
	LineEnding string

	APIURL               string
	Namespace            string
	Format               string
	AccessToken          string
	AccessTokenExpiresAt int64
	RefreshToken         string
	// TrustedCAFile string
	// InsecureSkipTLSVerify string
}

func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		shellCfg := shellConfig{
			APIURL:               cli.Config.APIUrl(),
			Namespace:            cli.Config.Namespace(),
			Format:               cli.Config.Format(),
			AccessToken:          cli.Config.Tokens().Access,
			AccessTokenExpiresAt: cli.Config.Tokens().ExpiresAt,
			RefreshToken:         cli.Config.Tokens().Refresh,
			// TrustedCAFile  cli.Config.TrustedCAFile(),
			// InsecureSkipTLSVerify: cli.Config.InsecureSkipTLSVerify(),
		}

		// Get the user shell
		shell := shell()

		// Determine if the shell flag was passed to override the shell to use
		shellFlag, err := cmd.Flags().GetString(shellFlag)
		if err != nil {
			return err
		}
		if shellFlag != "" {
			shell = shellFlag
		}

		switch shell {
		case "cmd":
			shellCfg.Prefix = "SET "
			shellCfg.Delimiter = "="
			shellCfg.LineEnding = "\n"
		case "powershell":
			shellCfg.Prefix = "$Env:"
			shellCfg.Delimiter = "="
			shellCfg.LineEnding = "\"\n"
		default: // bash
			shellCfg.Prefix = "export "
			shellCfg.Delimiter = "=\""
			shellCfg.LineEnding = "\"\n"
		}

		t := template.New("envConfig")
		tmpl, err := t.Parse(envTmpl)
		if err != nil {
			return err
		}

		return tmpl.Execute(os.Stdout, shellCfg)
	}
}

func refreshAccessToken(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		tokens, err := cli.Client.RefreshAccessToken(cli.Config.Tokens().Refresh)
		if err != nil {
			return fmt.Errorf(
				"failed to request new refresh token; client returned '%s'",
				err,
			)
		}

		// Write new tokens to disk
		return cli.Config.SaveTokens(tokens)
	}
}

func shell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return ""
	}

	return filepath.Base(shell)
}
