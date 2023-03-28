package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/types"
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

// PrintResourceJSON takes a record(s) and Resource, converts the record to
// human-readable JSON (pretty-prints), wraps that JSON using types.Wrapper, and
// then prints the result to the given writer. Unescapes any &, <, or >
// characters it finds.
func PrintResourceJSON(r corev3.Resource, wr io.Writer) error {
	w := types.WrapResource(r)

	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)

	if err := encoder.Encode(w); err != nil {
		return err
	}

	s := htmlReplacer.Replace(buf.String())
	_, err := fmt.Fprint(wr, s)
	return err
}

// PrintResourceListJSON takes a resource list and an io.Writer, converts the
// record to human-readable JSON (pretty-prints), wraps that JSON using
// types.Wrapper, and then prints the result to the given writer. Unescapes
// any &, <, or > characters it finds.
func PrintResourceListJSON(r []corev3.Resource, io io.Writer) error {
	for _, res := range r {
		err := PrintResourceJSON(res, io)
		if err != nil {
			return err
		}
	}
	return nil
}
