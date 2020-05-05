package token

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types/dynamic"
)

// Substitution evaluates the input template, that possibly contains
// tokens, with the provided data object and returns a slice of bytes
// representing the result along with any error encountered
func Substitution(data, input interface{}) ([]byte, error) {
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

// SubstituteAsset performs token substitution on an asset with the provided
// entity
func SubstituteAsset(asset *corev2.Asset, entity *corev2.Entity) error {
	// While we still validate an asset SHA512 value on creation/updates, we will
	// want to make sure it cannot be substituted for security reasons
	sha := asset.Sha512

	// Extract the extended attributes from the entity and combine them at the
	// top-level so they can be easily accessed using token substitution
	synthesizedEntity := dynamic.Synthesize(entity)

	// Substitute tokens within the asset with the synthesized entity
	bytes, err := Substitution(synthesizedEntity, asset)
	if err != nil {
		return err
	}

	// Unmarshal the asset obtained after the token substitution back into the
	// asset struct
	if err := json.Unmarshal(bytes, asset); err != nil {
		return fmt.Errorf("could not unmarshal the asset: %s", err)
	}

	// Set back the orginal SHA512 value
	asset.Sha512 = sha

	return nil
}

// SubstituteCheck performs token substitution on a check before its execution
// with the provided entity
func SubstituteCheck(check *corev2.CheckConfig, entity *corev2.Entity) error {
	// Extract the extended attributes from the entity and combine them at the
	// top-level so they can be easily accessed using token substitution
	synthesizedEntity := dynamic.Synthesize(entity)

	// Substitute tokens within the check configuration with the synthesized
	// entity
	bytes, err := Substitution(synthesizedEntity, check)
	if err != nil {
		return err
	}

	// Unmarshal the check configuration obtained after the token substitution
	// back into the check config struct
	if err := json.Unmarshal(bytes, check); err != nil {
		return fmt.Errorf("could not unmarshal the check: %s", err)
	}

	return nil
}

// SubstituteHook performs token substitution on a hook configuration with the
// provided entity
func SubstituteHook(hook *corev2.HookConfig, entity *corev2.Entity) error {
	// Extract the extended attributes from the entity and combine them at the
	// top-level so they can be easily accessed using token substitution
	synthesizedEntity := dynamic.Synthesize(entity)

	// Substitute tokens within the check configuration with the synthesized
	// entity
	bytes, err := Substitution(synthesizedEntity, hook)
	if err != nil {
		return err
	}

	// Unmarshal the check configuration obtained after the token substitution
	// back into the check config struct
	if err := json.Unmarshal(bytes, hook); err != nil {
		return fmt.Errorf("could not unmarshal the hook config: %s", err)
	}

	return nil
}

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
