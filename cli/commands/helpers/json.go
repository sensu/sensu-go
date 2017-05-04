package helpers

import (
	"encoding/json"
	"fmt"
	"io"
)

func PrintJSON(r interface{}, io io.Writer) {
	result, _ := json.MarshalIndent(r, "", "  ")
	fmt.Fprintf(io, "%s\n", result)
}
