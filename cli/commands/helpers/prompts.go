package helpers

import (
	"fmt"

	"github.com/AlecAivazis/survey"
	"github.com/sensu/sensu-go/cli/elements/globals"
)

// ConfirmDelete confirm a deletion operation before it is completed.
func ConfirmDelete(name string) bool {
	confirm := &ConfirmDestructiveOp{
		Type: "resource",
		Op:   "delete",
	}
	ok, _ := confirm.Ask(name)
	return ok
}

// ConfirmDestructiveOp presents prompt for a destructive operation.
type ConfirmDestructiveOp struct {
	Type string
	Op   string
}

// Ask presents prompt confirming a destructive operation.
func (c *ConfirmDestructiveOp) Ask(name string) (bool, error) {
	question := globals.TitleStyle("Are you sure you would like to ") +
		globals.CTATextStyle(c.Op) +
		globals.TitleStyle(fmt.Sprintf(" %s '", c.Type)) +
		globals.PrimaryTextStyle(name) +
		globals.TitleStyle("'?")

	confirm := &Confirm{
		Message: question,
		Default: false,
	}
	return confirm.Ask()
}

// Confirm an operation before it is completed
type Confirm struct {
	Message string
	Default bool
}

// Ask executes confirmation of operation
func (c *Confirm) Ask() (bool, error) {
	prompt := &survey.Confirm{
		Message: c.Message,
		Default: c.Default,
	}

	confirmation := false
	err := survey.AskOne(prompt, &confirmation, nil)
	if err != nil {
		return false, err
	}

	return confirmation, nil
}
