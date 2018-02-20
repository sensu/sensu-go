package asset

import (
	"errors"
	"fmt"
	"strings"

	"github.com/AlecAivazis/survey"
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
	exec := &CreateExecutor{
		Client: cli.Client,
		Org:    cli.Config.Organization(),
	}
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
		RunE: exec.Run,
	}

	_ = cmd.Flags().StringP("sha512", "", "", "SHA-512 checksum of the asset's archive")
	_ = cmd.Flags().StringP("url", "u", "", "the URL of the asset")
	_ = cmd.Flags().StringSliceP("metadata", "m", []string{}, "metadata associated with asset")
	_ = cmd.Flags().StringSlice("filter", []string{}, "queries used by an entity to determine if it should include the asset")

	helpers.AddInteractiveFlag(cmd.Flags())
	return cmd
}

// CreateExecutor executes create asset command
type CreateExecutor struct {
	Client client.APIClient
	Org    string
}

// Run runs the command given arguments
func (exePtr *CreateExecutor) Run(cmd *cobra.Command, args []string) error {
	if len(args) > 1 {
		_ = cmd.Help()
		return errors.New("invalid argument(s) received")
	}

	cfg := ConfigureAsset{
		Flags: cmd.Flags(),
		Args:  args,
		Org:   exePtr.Org,
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

	fmt.Fprint(cmd.OutOrStdout(), "OK")
	return nil
}

// ConfigureAsset given details configures a new asset
type ConfigureAsset struct {
	Flags *pflag.FlagSet
	Args  []string
	Org   string

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
		cfgPtr.setOrg()
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
			Name: "org",
			Prompt: &survey.Input{
				Message: "Organization:",
				Default: cfgPtr.Org,
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
	cfgPtr.setOrg()
	cfgPtr.setSha512()
	cfgPtr.setURL()
	cfgPtr.setMeta()
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

func (cfgPtr *ConfigureAsset) setOrg() {
	if len(cfgPtr.Org) == 0 {
		cfgPtr.addError(errors.New("organization name cannot be blank"))
	}

	cfgPtr.cfg.Org = cfgPtr.Org
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

func (cfgPtr *ConfigureAsset) setMeta() {
	if meta, err := cfgPtr.Flags.GetStringSlice("metadata"); err != nil {
		panic(err)
	} else {
		err = cfgPtr.cfg.SetMeta(meta)
		cfgPtr.addError(err)
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
	Name    string
	Org     string
	Sha512  string
	URL     string
	Meta    map[string]string
	Filters string
}

// SetMeta sets metadata given values
func (cfgPtr *Config) SetMeta(metadata []string) error {
	cfgPtr.Meta = make(map[string]string, len(metadata))
	for _, meta := range metadata {
		// TODO(james): naive
		splitMeta := strings.SplitAfterN(meta, ":", 2)

		if len(splitMeta) == 2 {
			key := strings.TrimSpace(strings.TrimRight(splitMeta[0], ":"))
			val := strings.TrimSpace(splitMeta[1])
			cfgPtr.Meta[key] = val
		} else {
			return fmt.Errorf(
				"Metadata value '%s' appears invalid;"+
					"should be in format 'KEY: VALUE'.",
				splitMeta,
			)
		}
	}
	return nil
}

// Copy applies configured details to given asset
func (cfgPtr *Config) Copy(asset *types.Asset) {
	asset.Name = cfgPtr.Name
	asset.Organization = cfgPtr.Org
	asset.Sha512 = cfgPtr.Sha512
	asset.URL = cfgPtr.URL
	asset.Metadata = cfgPtr.Meta
	asset.Filters = helpers.SafeSplitCSV(cfgPtr.Filters)
}
