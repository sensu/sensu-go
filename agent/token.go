package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

// TokenSubstitution evaluates the input template, that possibly contains
// tokens, with the provided data object and returns a slice of bytes
// representing the result along with any error encountered
func TokenSubstitution(data, input interface{}) ([]byte, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the provided template: %s", err)
	}

	var rawMap map[string]*json.RawMessage

	if err := json.Unmarshal(inputBytes, &rawMap); err != nil {
		return nil, err
	}

	for k, v := range rawMap {
		if v == nil || len(*v) == 0 || (*v)[0] != '"' {
			// null value, or not a string
			continue
		}
		tmpl := template.New(k)
		tmpl.Funcs(funcMap())

		var value string

		if err := json.Unmarshal([]byte(*v), &value); err != nil {
			return nil, fmt.Errorf("parsing %s: %s", k, err)
		}

		var err error
		tmpl, err = tmpl.Parse(value)
		if err != nil {
			return nil, fmt.Errorf("could not parse the template: %s", err)
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, data)
		if err != nil {
			return nil, fmt.Errorf("could not execute the template: %s", err)
		}

		// Verify if the output contains the "<no value>" string, indicating that a
		// token was not properly substituted. If so, re-execute the template but this
		// time with "missingkey=error" option so we get the actual token that was
		// unmatched. For reference, this option can't be added by default otherwise
		// the default values (defaultFunc) couldn't work
		if strings.Contains(buf.String(), "<no value>") {
			tmpl.Option("missingkey=error")

			if err = tmpl.Execute(&buf, data); err == nil {
				return nil, fmt.Errorf("%s: unmatched token: found an undefined value but could not identify the token", k)
			}

			return nil, fmt.Errorf("%s: unmatched token: %s", k, err)
		}

		templated, _ := json.Marshal(buf.String())

		rawMap[k] = (*json.RawMessage)(&templated)
	}

	return json.Marshal(rawMap)
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
		if v[1] == nil {
			return v[0]
		}
		return v[1]
	}
	return nil
}
