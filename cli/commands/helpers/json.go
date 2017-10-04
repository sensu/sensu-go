package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// PrintJSON takes a record(s) and an io.Writer, converts the record to human-
// readable JSON (prrtty-prints), and then prints the result to the given
// writer. SetEscapeHTML is necessary to avoid printing &, <, and > as unicode
// values.
func PrintJSON(r interface{}, io io.Writer) error {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(r); err != nil {
		return err
	}

	fmt.Fprintf(io, "%s\n", buf.Bytes())
	return nil
}
