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
	Type string                 `yaml:"type"`
	Spec map[string]interface{} `yaml:"spec"`
}

func wrapResource(r types.Resource) wrapper {
	wrapped := types.WrapResource(r)
	w := wrapper{
		Type: wrapped.Type,
		Spec: toMap(wrapped.Value),
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
	defer func() {
		if err == nil {
			err = enc.Close()
		}
	}()
	if resources, ok := v.([]types.Resource); ok {
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
