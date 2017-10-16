package user

import (
	"errors"
	"fmt"

	"github.com/AlecAivazis/survey"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/cli/commands/hooks"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

const (
	passwordsDoNotMatchError = "given passwords do not match"
	badCurrentPasswordError  = "given password did not match the one on file"
)

type passwordPromptInput struct {
	New     string `survey:"new-password"`
	Confirm string `survey:"confirm-password"`
}

// SetPasswordCommand adds command that allows user to create new users
func SetPasswordCommand(cli *cli.SensuCli) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "change-password USERNAME",
		Short:        "change password for given user",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var username string
			var promptForCurrentPassword bool

			// Retrieve current username from JWT
			currentUsername := getCurrentUsername(cli.Config)

			// If no username is given we user the current user's name
			if len(args) > 0 {
				username = args[0]
			} else {
				username = currentUsername
				promptForCurrentPassword = true
			}

			// As a precaution ask for the current user's password
			if promptForCurrentPassword {
				if err := verifyExistingPassword(currentUsername, cli); err != nil {
					return errors.New(badCurrentPasswordError)
				}
			}

			// Prompt user for new password
			inputs := passwordPromptInput{}
			if err := administerQuestionnaire(&inputs); err != nil {
				return err
			}

			// Validate new password
			if err := validateInput(&inputs); err != nil {
				return err
			}

			// Update password
			err := cli.Client.UpdatePassword(username, inputs.New)
			if err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Updated")
			return nil
		},
		Annotations: map[string]string{
			// We want to be able to run this command regardless of whether the CLI
			// has been configured.
			hooks.ConfigurationRequirement: hooks.ConfigurationNotRequired,
		},
	}

	return cmd
}

func getCurrentUsername(cfg config.Config) string {
	accessToken := cfg.Tokens().Access
	token, _ := jwt.ParseWithClaims(accessToken, &types.Claims{}, nil)
	claims := token.Claims.(*types.Claims)
	return claims.StandardClaims.Subject
}

func verifyExistingPassword(username string, cli *cli.SensuCli) error {
	inputs := struct{ Password string }{}
	qs := []*survey.Question{
		{
			Name:     "password",
			Prompt:   &survey.Password{Message: "Current Password:"},
			Validate: survey.Required,
		},
	}

	// Get password
	if err := survey.Ask(qs, &inputs); err != nil {
		return err
	}

	// Attempt to authenticate
	_, err := cli.Client.CreateAccessToken(cli.Config.APIUrl(), username, inputs.Password)
	if err != nil {
		return err
	}

	return nil
}

func administerQuestionnaire(inputs *passwordPromptInput) error {
	qs := []*survey.Question{
		{
			Name:     "new-password",
			Prompt:   &survey.Password{Message: "Password:"},
			Validate: survey.Required,
		},
		{
			Name:     "confirm-password",
			Prompt:   &survey.Password{Message: "Confirm:"},
			Validate: survey.Required,
		},
	}

	return survey.Ask(qs, inputs)
}

func validateInput(inputs *passwordPromptInput) error {
	if inputs.New != inputs.Confirm {
		return errors.New(passwordsDoNotMatchError)
	}

	user := types.User{Password: inputs.New}
	return user.ValidatePassword()
}
