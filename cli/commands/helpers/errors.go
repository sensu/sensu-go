package helpers

import "errors"

// JoinErrors joins multiple errors messages. Useful when
// you want the CLI to display more than one error message.
//
// eg.
//
//   JoinErrors("Validation: ", []error{errors.New("a"), errors.New("b")})
//   "Validation: a, and b."
func JoinErrors(prelude string, errs []error) error {
	out := prelude + " "
	lastElem := len(errs) - 1

	for i, err := range errs {
		var seperator string
		switch i {
		case 1:
			seperator = ""
		case lastElem:
			seperator = ", and "
		default:
			seperator = ", "
		}

		out += seperator + err.Error() + "."
	}

	return errors.New(out)
}
