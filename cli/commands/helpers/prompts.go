package helpers

import (
	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/elements/globals"
)

// ConfirmDelete confirm a deletion operation before it is completed.
func ConfirmDelete(name string) bool {
	question := globals.TitleStyle("Are you sure you would like to ") + globals.CTATextStyle("delete") + globals.TitleStyle(" resource '") + globals.PrimaryTextStyle(name) + globals.TitleStyle("'?")

	confirmation := false
	prompt := &survey.Confirm{
		Message: question,
		Default: false,
	}
	err := survey.AskOne(prompt, &confirmation, nil)
	if err != nil {
		return false
	}

	return confirmation
}
