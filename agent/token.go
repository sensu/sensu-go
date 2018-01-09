package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
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

	inputString := strings.Replace(string(inputBytes), `\"`, `"`, -1)

	tmpl := template.New("")
	tmpl.Funcs(funcMap())

	// An error should be returned if a token can't be properly substituted
	tmpl.Option("missingkey=error")

	tmpl, err = tmpl.Parse(inputString)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template: %s", err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("could not execute the template: %s", err.Error())
	}

	return buf.Bytes(), nil
}

// funcMap defines the available custom functions in templates
func funcMap() template.FuncMap {
	return template.FuncMap{
		"default": defaultFunc,
	}
}

// defaultFunc receives v, a slice of interfaces, which length range between one
// and two arguments, depending on whether the token has a corresponding field.
// The first argument always represents the default value, while the optional
// second argument represent the value of the token if it was properly
// substitued, in which case we should return that value instead of the default
func defaultFunc(v ...interface{}) interface{} {
	if len(v) == 1 {
		return v[0]
	} else if len(v) == 2 {
		return v[1]
	}
	return nil
}
