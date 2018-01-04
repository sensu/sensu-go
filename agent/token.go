package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
)

// tokenSubstitution evaluates the input template, that possibly contains
// tokens, with the provided data object and returns a slice of bytes
// representing the result along with any error encountered
func tokenSubstitution(data, input interface{}) ([]byte, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the provided template: %s", err.Error())
	}

	tmpl := template.New("")
	tmpl, err = tmpl.Parse(string(inputBytes))

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("could not execute the template: %s", err.Error())
	}

	return buf.Bytes(), nil
}
