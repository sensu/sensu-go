package v2

import (
	"errors"
	"strings"
)

func validateVar(v string) error {
	parts := strings.SplitN(v, "=", 2)
	if len(parts) != 2 {
		return errors.New("environment variables must be of the form FOO=BAR")
	}
	if len(parts[0]) == 0 || len(parts[1]) == 0 {
		return errors.New("environment variables must be of the form FOO=BAR")
	}
	return nil
}

// ValidateEnvVars ensures that all the environment variables are well-formed.
// Vars should be of the form FOO=BAR.
func ValidateEnvVars(vars []string) error {
	for _, v := range vars {
		if err := validateVar(v); err != nil {
			return err
		}
	}
	return nil
}

// EnvVarsToMap converts a list of FOO=BAR key-value pairs into a map.
func EnvVarsToMap(vars []string) map[string]string {
	result := make(map[string]string, len(vars))
	for _, v := range vars {
		parts := strings.SplitN(v, "=", 2)
		if len(parts) == 1 {
			continue
		}
		result[parts[0]] = parts[1]
	}
	return result
}
