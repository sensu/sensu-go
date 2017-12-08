package helpers

import (
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/sensu/sensu-go/cli/elements/globals"
)

// ConfirmDelete confirm a deletion operation before it is completed.
func ConfirmDelete(name string, stdout io.Writer) bool {
	confirmation := strings.ToUpper(name)

	title := globals.TitleStyle("Are you sure you would like to ") + globals.CTATextStyle("delete") + globals.TitleStyle(" resource '") + globals.PrimaryTextStyle(name) + globals.TitleStyle("'?")
	question := "Enter '" + globals.PrimaryTextStyle(confirmation) + "' to confirm."

	fmt.Fprintf(stdout, "%s\n\n%s\n", title, question)

	// NOTE: configured properly NewEx should never return an error
	rl, _ := readline.NewEx(&readline.Config{Prompt: "> "})
	defer func() {
		_ = rl.Close()
	}()

	line, _ := rl.Readline()
	return confirmation == line
}
