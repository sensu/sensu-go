package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

func substituteToken(key string, data interface{}, message *json.RawMessage) (*json.RawMessage, error) {
	if message == nil {
		return nil, nil
	}
	if len(*message) == 0 {
		return message, nil
	}
	switch (*message)[0] {
	case '"':
		return substituteString(key, data, message)
	case '[':
		return substituteArray(key, data, message)
	case '{':
		var object map[string]*json.RawMessage
		if err := json.Unmarshal([]byte(*message), &object); err != nil {
			return nil, fmt.Errorf("couldn't evaluate template for %s: %s (object)", key, err)
		}
		for k, v := range object {
			value, err := substituteToken(k, data, v)
			if err != nil {
				return nil, err
			}
			object[k] = value
		}
		b, _ := json.Marshal(object)
		return (*json.RawMessage)(&b), nil
	default:
		return message, nil
	}
}

func substituteString(key string, data interface{}, message *json.RawMessage) (*json.RawMessage, error) {
	var t string
	if err := json.Unmarshal([]byte(*message), &t); err != nil {
		return nil, fmt.Errorf("couldn't evaluate template for %s: %s (string)", key, err)
	}

	tmpl := template.New(key)
	tmpl.Funcs(funcMap())

	var err error
	tmpl, err = tmpl.Parse(t)
	if err != nil {
		return nil, fmt.Errorf("%s: could not parse the template: %s", key, err)
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
			return nil, fmt.Errorf("%s: unmatched token: found an undefined value but could not identify the token", key)
		}

		return nil, fmt.Errorf("%s: unmatched token: %s", key, err)
	}

	templated, _ := json.Marshal(buf.String())

	return (*json.RawMessage)(&templated), nil
}

func substituteArray(key string, data interface{}, message *json.RawMessage) (*json.RawMessage, error) {
	var messages []*json.RawMessage
	if err := json.Unmarshal([]byte(*message), &messages); err != nil {
		return nil, fmt.Errorf("couldn't evaluate template for %s: %s (array)", key, err)
	}

	for i := range messages {
		templated, err := substituteToken(key, data, messages[i])
		if err != nil {
			return nil, fmt.Errorf("couldn't evaluate template for %s: %s (array %d)", key, err, i)
		}
		messages[i] = templated
	}

	b, _ := json.Marshal(messages)

	return (*json.RawMessage)(&b), nil
}

// TokenSubstitution evaluates the input template, that possibly contains
// tokens, with the provided data object and returns a slice of bytes
// representing the result along with any error encountered
func TokenSubstitution(data, input interface{}) ([]byte, error) {
	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal the provided template: %s", err)
	}

	rawMessage, err := substituteToken("", data, (*json.RawMessage)(&inputBytes))
	if err != nil {
		return nil, err
	}

	return []byte(*rawMessage), nil
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
