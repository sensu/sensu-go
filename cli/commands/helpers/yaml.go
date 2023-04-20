package helpers

import (
	"io"

	v2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/types"
	yaml "gopkg.in/yaml.v2"
)

// PrintYAML serializes the value v to yaml and writes the result to w.
func PrintYAML(v interface{}, w io.Writer) (err error) {
	enc := yaml.NewEncoder(w)
	var close = true
	defer func() {
		if err == nil && close {
			err = enc.Close()
		}
	}()
	if resources, ok := v.([]corev3.Resource); ok {
		if len(resources) == 0 {
			close = false
			return nil
		}
		for _, r := range resources {
			i := types.WrapResource(r)
			if err := enc.Encode(i); err != nil {
				return err
			}
		}
		return nil
	}
	if r, ok := v.(v2.Resource); ok {
		v = types.WrapResource(r)
	}

	return enc.Encode(v)
}
