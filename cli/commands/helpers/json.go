package helpers

import (
	"encoding/json"
	"fmt"
	"io"
)

// PrintJSON takes a record(s) and an io.Writer, converts the record to human-
// readable JSON (prrtty-prints), and then prints the result to the given
// writer.
func PrintJSON(r interface{}, io io.Writer) {
	result, _ := json.MarshalIndent(r, "", "  ")
	fmt.Fprintf(io, "%s\n", result)
}
