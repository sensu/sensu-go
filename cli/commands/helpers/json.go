package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var htmlReplacer = strings.NewReplacer(`\u0026`, "&", `\u003c`, "<", `\u003e`, ">")

// PrintJSON takes a record(s) and an io.Writer, converts the record to human-
// readable JSON (pretty-prints), and then prints the result to the given
// writer. Unescapes any &, <, or > characters it finds.
func PrintJSON(r interface{}, io io.Writer) error {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(r); err != nil {
		return err
	}

	s := htmlReplacer.Replace(buf.String())
	_, err := fmt.Fprintln(io, s)
	return err
}
