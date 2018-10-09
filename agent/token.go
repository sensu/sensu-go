package agent

import (
	"bytes"
	"encoding/json"
	"errors"
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
		return nil, fmt.Errorf("could not marshal the provided template: %s", err.Error())
	}

	// replace special character \" with " only if contained within {{ }}
	inputString := string(inputBytes)
	for i := range inputString {
		if i < len(inputString)-1 {
			if string(inputString[i]) == "{" && string(inputString[i+1]) == "{" {
				rightSlice := inputString[i+2:]
				inner := strings.Split(rightSlice, "}}")[0]
				innerParsed := strings.Replace(inner, "\\\"", "\"", -1)
				inputString = strings.Replace(inputString, inner, innerParsed, -1)
			}
		}
	}

	tmpl := template.New("")
	tmpl.Funcs(funcMap())

	tmpl, err = tmpl.Parse(inputString)
	if err != nil {
		return nil, fmt.Errorf("could not parse the template: %s", err.Error())
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return nil, fmt.Errorf("could not execute the template: %s", err.Error())
	}

	// Verify if the output contains the "<no value>" string, indicating that a
	// token was not properly substituted. If so, re-execute the template but this
	// time with "missingkey=error" option so we get the actual token that was
	// unmatched. For reference, this option can't be added by default otherwise
	// the default values (defaultFunc) couldn't work
	if strings.Contains(buf.String(), "<no value>") {
		tmpl.Option("missingkey=error")

		if err = tmpl.Execute(&buf, data); err == nil {
			return nil, errors.New("unmatched token: found an undefined value but could not identify the token")
		}

		return nil, fmt.Errorf("unmatched token: %s", err.Error())
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
		if v[1] == nil {
			return v[0]
		}
		return v[1]
	}
	return nil
}
