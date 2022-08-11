package helpers

import (
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	"github.com/sensu/sensu-go/cli/elements/globals"
)

// ConfirmDelete confirm a deletion operation before it is completed.
func ConfirmDelete(name string, opts ...survey.AskOpt) bool {
	confirm := &ConfirmDestructiveOp{
		Type:    "resource",
		Op:      "delete",
		AskOpts: opts,
	}
	ok, _ := confirm.Ask(name)
	return ok
}

// ConfirmDelete confirm a deletion operation before it is completed.
func ConfirmDeleteWithOpts(name string, opts ...survey.AskOpt) bool {
	confirm := &ConfirmDestructiveOp{
		Type:    "resource",
		Op:      "delete",
		AskOpts: opts,
	}
	ok, _ := confirm.Ask(name)
	return ok
}

// ConfirmDeleteResource confirm a deletion operation before it is completed.
func ConfirmDeleteResource(name string, resourceType string) bool {
	return ConfirmDeleteResourceWithOpts(name, resourceType, nil)
}

// ConfirmDeleteResourceWithOpts confirm a deletion operation before it is
// completed. It accepts a list of survey.AskOpt for customizing behaviour.
func ConfirmDeleteResourceWithOpts(name string, resourceType string, opts ...survey.AskOpt) bool {
	confirm := &ConfirmDestructiveOp{
		Type:    fmt.Sprintf("%s resource", resourceType),
		Op:      "delete",
		AskOpts: opts,
	}
	ok, _ := confirm.Ask(name)
	return ok
}

// ConfirmDestructiveOp presents prompt for a destructive operation.
type ConfirmDestructiveOp struct {
	Type    string
	Op      string
	AskOpts []survey.AskOpt
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
		AskOpts: c.AskOpts,
	}
	return confirm.Ask()
}

// Confirm an operation before it is completed
type Confirm struct {
	Message string
	Default bool
	AskOpts []survey.AskOpt
}

// Ask executes confirmation of operation
func (c *Confirm) Ask() (bool, error) {
	prompt := &survey.Confirm{
		Message: c.Message,
		Default: c.Default,
	}

	confirmation := false
	if err := survey.AskOne(prompt, &confirmation, c.AskOpts...); err != nil {
		return false, err
	}

	return confirmation, nil
}

// ConfirmOptOut confirm an opt-out operation before it is completed.
func ConfirmOptOut() bool {
	c := &Confirm{
		Message: "Are you sure you would like to opt-out of tessen? We'd hate to see you go!",
		Default: false,
	}
	ok, _ := c.Ask()
	return ok
}
