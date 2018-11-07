package helpers

import (
	"errors"
	"fmt"
	"strings"
)

// VerifyName ensures that (only) a name was provided in the arguments
func VerifyName(args []string) error {
	if len(args) == 0 {
		return errors.New("a name is required")
	}
	if len(args) > 1 {
		return fmt.Errorf("too many arguments supplied: %s", strings.Join(args, ", "))
	}
	return nil
}
