package env

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"text/template"

	"github.com/sensu/sensu-go/cli"
	"github.com/spf13/cobra"
)

const (
	envTmpl = `{{ .Prefix }}SENSU_API_URL{{ .Delimiter }}{{ .APIURL }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_NAMESPACE{{ .Delimiter }}{{ .Namespace }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_FORMAT{{ .Delimiter }}{{ .Format }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_API_KEY{{ .Delimiter }}{{ .APIKey }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_ACCESS_TOKEN{{ .Delimiter }}{{ .AccessToken }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_ACCESS_TOKEN_EXPIRES_AT{{ .Delimiter }}{{ .AccessTokenExpiresAt }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_REFRESH_TOKEN{{ .Delimiter }}{{ .RefreshToken }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_TRUSTED_CA_FILE{{ .Delimiter }}{{ .TrustedCAFile }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_INSECURE_SKIP_TLS_VERIFY{{ .Delimiter }}{{ .InsecureSkipTLSVerify }}{{ .LineEnding }}` +
		`{{ .Prefix }}SENSU_TIMEOUT{{ .Delimiter }}{{ .Timeout }}{{ .LineEnding }}` +
		`{{ .UsageHint }}`

	shellFlag = "shell"
)

// Command display the commands to set up the environment used by sensuctl
func Command(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "env",
		Short:  "Display the commands to set up the environment used by sensuctl",
		PreRun: refreshAccessToken(cli),
		RunE:   execute(cli),
	}

	_ = cmd.Flags().StringP(shellFlag, "", "",
		fmt.Sprintf(
			`force environment to be configured for a specified shell ("%s"|"%s"|"%s")`,
			"bash", "cmd", "powershell",
		))

	return cmd
}

type shellConfig struct {
	args      []string
	userShell string

	Prefix     string
	Delimiter  string
	LineEnding string

	APIURL                string
	Namespace             string
	Format                string
	APIKey                string
	AccessToken           string
	AccessTokenExpiresAt  int64
	RefreshToken          string
	TrustedCAFile         string
	InsecureSkipTLSVerify string
	Timeout               string
}

func (s shellConfig) UsageHint() string {
	cmd := ""
	comment := "#"
	commandLine := strings.Join(s.args, " ")

	switch s.userShell {
	case "cmd":
		cmd = fmt.Sprintf("\t@FOR /f \"tokens=*\" %%i IN ('%s') DO @%%i", commandLine)
		comment = "REM"
	case "powershell":
		cmd = fmt.Sprintf("& %s | Invoke-Expression", commandLine)
	default:
		cmd = fmt.Sprintf("eval $(%s)", commandLine)
	}

	return fmt.Sprintf("%s Run this command to configure your shell: \n%s %s\n", comment, comment, cmd)
}

// execute contains the actual logic for displaying the environment
func execute(cli *cli.SensuCli) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		shellCfg := shellConfig{
			args:                  os.Args,
			APIURL:                cli.Config.APIUrl(),
			Namespace:             cli.Config.Namespace(),
			Format:                cli.Config.Format(),
			APIKey:                cli.Config.APIKey(),
			AccessToken:           cli.Config.Tokens().Access,
			AccessTokenExpiresAt:  cli.Config.Tokens().ExpiresAt,
			RefreshToken:          cli.Config.Tokens().Refresh,
			TrustedCAFile:         cli.Config.TrustedCAFile(),
			InsecureSkipTLSVerify: strconv.FormatBool(cli.Config.InsecureSkipTLSVerify()),
			Timeout:               cli.Config.Timeout().String(),
		}

		// Get the user shell
		shellCfg.userShell = shell()

		// Determine if the shell flag was passed to override the shell to use
		shellFlag, err := cmd.Flags().GetString(shellFlag)
		if err != nil {
			return err
		}
		if shellFlag != "" {
			shellCfg.userShell = shellFlag
		}

		switch shellCfg.userShell {
		case "cmd":
			shellCfg.Prefix = "SET "
			shellCfg.Delimiter = "="
			shellCfg.LineEnding = "\n"
		case "powershell":
			shellCfg.Prefix = "$Env:"
			shellCfg.Delimiter = " = \""
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

		return tmpl.Execute(cmd.OutOrStdout(), shellCfg)
	}
}

// refreshAccessToken attempts to silently refresh the access token
func refreshAccessToken(cli *cli.SensuCli) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		tokens, err := cli.Client.RefreshAccessToken(cli.Config.Tokens())
		if err != nil {
			return
		}

		// Write new tokens to disk
		_ = cli.Config.SaveTokens(tokens)
	}
}

// shell attempts to discover the shell currently used
func shell() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		// Default to powershell for now when running on Windows
		if runtime.GOOS == "windows" {
			return "powershell"
		}
		return ""
	}

	return filepath.Base(shell)
}
