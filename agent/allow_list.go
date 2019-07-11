package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v2"
)

type allowList struct {
	Exec      string   `yaml:"exec" json:"exec"`
	Args      []string `yaml:"args" json:"args"`
	Sha512    string   `yaml:"sha512" json:"sha512"`
	EnableEnv bool     `yaml:"enable_env" json:"enable_env"`
}

func readAllowList(path string, readBytes func(string) ([]byte, error)) ([]allowList, error) {
	var allowList []allowList
	if path == "" {
		return allowList, nil
	}
	unmarshalFuncs := map[string]func(in []byte, out interface{}) error{
		".yaml": yaml.Unmarshal,
		".yml":  yaml.Unmarshal,
		".json": json.Unmarshal,
	}
	for ext, f := range unmarshalFuncs {
		if strings.Contains(path, ext) {
			bytes, err := readBytes(path)
			if err != nil {
				return nil, err
			}
			err = f(bytes, &allowList)
			if err != nil {
				return nil, err
			}
			for _, al := range allowList {
				err = al.validate()
				if err != nil {
					return nil, err
				}
			}
			return allowList, nil
		}
	}

	return nil, fmt.Errorf("invalid file extension")
}

// validate returns an error if the allowList contains invalid values.
func (al *allowList) validate() error {
	if al.Exec == "" {
		return errors.New("exec cannot be empty")
	}

	if len(al.Args) == 0 {
		return errors.New("args cannot be empty")
	}

	return nil
}

func (a *Agent) matchAllowList(command string) (allowList, bool) {
	for _, al := range a.allowList {
		remaining := command
		if strings.Contains(command, al.Exec) {
			remaining = strings.Replace(remaining, al.Exec, "", -1)
			for _, a := range al.Args {
				if strings.Contains(command, a) {
					remaining = strings.Replace(remaining, a, "", -1)
				}
			}
			if strings.TrimSpace(remaining) == "" {
				return al, true
			}
		}
	}
	return allowList{}, false
}
