package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/sensu/sensu-go/cli/client/config"
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

// PrintWrappedJSON takes a record(s) and Resource, converts the record to
// human-readable JSON (pretty-prints), wraps that JSON using types.Wrapper, and
// then prints the result to the given writer. Unescapes any &, <, or >
// characters it finds.
func PrintWrappedJSON(r types.Resource, wr io.Writer) error {
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

// PrintWrappedJSONList takes a resource list and an io.Writer, converts the
// record to human-readable JSON (pretty-prints), wraps that JSON using
// types.Wrapper, and then prints the result to the given writer. Unescapes
// any &, <, or > characters it finds.
func PrintWrappedJSONList(r []types.Resource, io io.Writer) error {
	for _, res := range r {
		err := PrintWrappedJSON(res, io)
		if err != nil {
			return err
		}
	}
	return nil
}

// PrintFormatted prints the provided interface in the specified format.
// flag overrides the cli config format if set
func PrintFormatted(flag string, format string, v interface{}, w io.Writer, printToList func(interface{}, io.Writer) error) error {
	if flag != "" {
		format = flag
	}
	switch format {
	case config.FormatJSON:
		return PrintJSON(v, w)
	case config.FormatWrappedJSON:
		r, ok := v.(types.Resource)
		if !ok {
			return fmt.Errorf("%t is not a Resource", v)
		}
		return PrintWrappedJSON(r, w)
	default:
		return printToList(v, w)
	}
}
