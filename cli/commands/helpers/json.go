package helpers

import (
	"encoding/json"
	"fmt"
	"os"
)

func PrintResultsToPrettyJSON(r interface{}) {
	result, _ := json.MarshalIndent(r, "", "  ")
	fmt.Fprintf(os.Stdout, "%s\n", result)
}
