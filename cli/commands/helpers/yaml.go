package helpers

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/sensu/sensu-go/types"
	yaml "github.com/sensu/yaml"
)

// wrapper is used to get the precise yaml encoding behaviour we want
type wrapper struct {
	Type       string                 `json:"type" yaml:"type"`
	APIVersion string                 `json:"api_version" yaml:"api_version"`
	ObjectMeta map[string]interface{} `json:"metadata" yaml:"metadata"`
	Spec       map[string]interface{} `json:"spec" yaml:"spec"`
}

func wrapResource(r types.Resource) wrapper {
	wrapped := types.WrapResource(r)
	value := toMap(wrapped.Value)
	delete(value, "metadata")
	w := wrapper{
		Type:       wrapped.Type,
		APIVersion: wrapped.APIVersion,
		ObjectMeta: toMap(wrapped.ObjectMeta),
		Spec:       value,
	}
	return w
}

// toMap produces a map from a struct by serializing it to JSON and then
// deserializing the JSON into a map. This is done to preserve business logic
// expressed in customer marshalers, and JSON struct tag semantics.
func toMap(v interface{}) map[string]interface{} {
	b, _ := json.Marshal(v)
	result := map[string]interface{}{}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	_ = dec.Decode(&result)
	return result
}

// PrintYAML serializes the value v to yaml and writes the result to w.
func PrintYAML(v interface{}, w io.Writer) (err error) {
	enc := yaml.NewEncoder(w)
	var close = true
	defer func() {
		if err == nil && close {
			err = enc.Close()
		}
	}()
	if resources, ok := v.([]types.Resource); ok {
		if len(resources) == 0 {
			close = false
			return nil
		}
		for _, r := range resources {
			if err := enc.Encode(wrapResource(r)); err != nil {
				return err
			}
		}
		return nil
	}
	if r, ok := v.(types.Resource); ok {
		v = wrapResource(r)
	}
	return enc.Encode(v)
}
