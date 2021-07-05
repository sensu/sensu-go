package asset

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/commands/flags"
	"github.com/sensu/sensu-go/cli/commands/helpers"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CreateCommand adds command that allows user to create new assets
func CreateCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [NAME]",
		Short: "create new assets",
		PreRun: func(cmd *cobra.Command, args []string) {
			isInteractive, _ := cmd.Flags().GetBool(flags.Interactive)
			if !isInteractive {
				// Mark flags are required for bash-completions
				_ = cmd.MarkFlagRequired("sha512")
				_ = cmd.MarkFlagRequired("url")
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			executor := &CreateExecutor{
				Client:    cli.Client,
				Namespace: cli.Config.Namespace(),
			}

			return executor.Run(cmd, args)
		},
	}

	_ = cmd.Flags().StringP("sha512", "", "", "SHA-512 checksum of the asset's archive")
	_ = cmd.Flags().StringP("url", "u", "", "the URL of the asset")
	_ = cmd.Flags().StringSlice("filter", []string{}, "queries used by an entity to determine if it should include the asset")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}

// CreateExecutor executes create asset command
type CreateExecutor struct {
	Client    client.APIClient
	Namespace string
}

// Run runs the command given arguments
func (exePtr *CreateExecutor) Run(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		_ = cmd.Help()
		return errors.New("invalid argument(s) received")
	}

	cfg := ConfigureAsset{
		Flags:     cmd.Flags(),
		Args:      args,
		Namespace: exePtr.Namespace,
	}

	asset, errs := cfg.Configure()
	if len(errs) > 0 {
		return helpers.JoinErrors("Bad inputs: ", errs)
	}

	if err := asset.Validate(); err != nil {
		return err
	}

	if err := exePtr.Client.CreateAsset(asset); err != nil {
		return err
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Created")
	return nil
}

// ConfigureAsset given details configures a new asset
type ConfigureAsset struct {
	Flags     *pflag.FlagSet
	Args      []string
	Namespace string

	cfg    Config
	errors []error
}

// Configure returns a new asset or returns error if arguments are invalid
func (cfgPtr *ConfigureAsset) Configure() (*types.Asset, []error) {
	isInteractive, _ := cfgPtr.Flags.GetBool(flags.Interactive)

	if len(cfgPtr.Args) == 1 {
		cfgPtr.cfg.Name = cfgPtr.Args[0]
	} else if len(cfgPtr.Args) > 1 {
		cfgPtr.addError(errors.New("too many arguments given"))
	}

	if isInteractive {
		cfgPtr.setNamespace()
		if err := cfgPtr.administerQuestionnaire(); err != nil {
			cfgPtr.addError(err)
		}
	} else {
		cfgPtr.configureFromFlags()
	}

	var asset types.Asset
	cfgPtr.cfg.Copy(&asset)

	return &asset, cfgPtr.errors
}

func (cfgPtr *ConfigureAsset) administerQuestionnaire() error {
	var qs = []*survey.Question{
		{
			Name: "name",
			Prompt: &survey.Input{
				Message: "Name:",
				Default: cfgPtr.cfg.Name,
			},
			Validate: survey.Required,
		},
		{
			Name: "namespace",
			Prompt: &survey.Input{
				Message: "Namespace:",
				Default: cfgPtr.Namespace,
			},
			Validate: survey.Required,
		},
		{
			Name:     "url",
			Prompt:   &survey.Input{Message: "URL:"},
			Validate: survey.Required,
		},
		{
			Name:     "sha512",
			Prompt:   &survey.Input{Message: "SHA-512 Checksum:"},
			Validate: survey.Required,
		},
		{
			Name:   "filters",
			Prompt: &survey.Input{Message: "Filters:"},
		},
	}

	return survey.Ask(qs, &cfgPtr.cfg)
}

func (cfgPtr *ConfigureAsset) configureFromFlags() {
	cfgPtr.setName()
	cfgPtr.setNamespace()
	cfgPtr.setSha512()
	cfgPtr.setURL()
	cfgPtr.setFilters()
}

func (cfgPtr *ConfigureAsset) setName() {
	args := cfgPtr.Args

	if len(args) == 1 {
		cfgPtr.cfg.Name = args[0]
	} else {
		cfgPtr.addError(errors.New("please provide a name"))
	}
}

func (cfgPtr *ConfigureAsset) setNamespace() {
	if len(cfgPtr.Namespace) == 0 {
		cfgPtr.addError(errors.New("namespace name cannot be blank"))
	}

	cfgPtr.cfg.Namespace = cfgPtr.Namespace
}

func (cfgPtr *ConfigureAsset) setSha512() {
	if sha512, err := cfgPtr.Flags.GetString("sha512"); err != nil {
		panic(err)
	} else {
		cfgPtr.cfg.Sha512 = sha512
	}
}

func (cfgPtr *ConfigureAsset) setURL() {
	if url, err := cfgPtr.Flags.GetString("url"); err != nil {
		panic(err)
	} else {
		cfgPtr.cfg.URL = url
	}
}

func (cfgPtr *ConfigureAsset) setFilters() {
	if filters, err := cfgPtr.Flags.GetStringSlice("filter"); err != nil {
		panic(err)
	} else {
		cfgPtr.cfg.Filters = strings.Join(filters, ",")
	}
}

func (cfgPtr *ConfigureAsset) addError(err error) {
	if err != nil {
		cfgPtr.errors = append(cfgPtr.errors, err)
	}
}

// Config represents configurable attributes of an asset
type Config struct {
	Name      string
	Namespace string
	Sha512    string
	URL       string
	Filters   string
}

// Copy applies configured details to given asset
func (cfgPtr *Config) Copy(asset *types.Asset) {
	asset.Name = cfgPtr.Name
	asset.Namespace = cfgPtr.Namespace
	asset.Sha512 = cfgPtr.Sha512
	asset.URL = cfgPtr.URL
	asset.Filters = helpers.SafeSplitCSV(cfgPtr.Filters)
}
