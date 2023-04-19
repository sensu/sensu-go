package request

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	corev3 "github.com/sensu/core/v3"
	types "github.com/sensu/core/v3/types"
	apitools "github.com/sensu/sensu-api-tools"
)

// Resource decodes the request body into the specified corev3.Resource type
func Resource[R corev3.Resource](r *http.Request) (R, error) {
	var payload R

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return payload, err
	}

	if err := validate[R](body); err != nil {
		return payload, err
	}

	var wrapper types.Wrapper
	if err := json.Unmarshal(body, &wrapper); err != nil {
		return payload, err
	}
	if _, ok := wrapper.Value.(R); !ok {
		return payload, fmt.Errorf("unexpected type described in the request body: expected %T, got %T", payload, wrapper.Value)
	}

	return wrapper.Value.(R), nil
}

func validate[R corev3.Resource](b []byte) error {
	var w rawWrapper
	if err := json.Unmarshal(b, &w); err != nil {
		return err
	}

	var missing []string
	if w.APIVersion == "" {
		missing = append(missing, "api_version")
	}
	if w.Type == "" {
		missing = append(missing, "type")
	}
	if len(w.ObjectMeta) == 0 {
		missing = append(missing, "metadata")
	}
	if len(w.Spec) == 0 {
		missing = append(missing, "spec")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required resource identifying fields: %v. could be using an unsupported legacy request format", missing)
	}

	if t, err := apitools.Resolve(w.APIVersion, w.Type); err != nil {
		return fmt.Errorf("error resolving the resource type described in the request body: %v", err)
	} else if _, ok := t.(R); !ok {
		var expected R
		return fmt.Errorf("type described in the request body unexpected: expected %T, got %T", expected, t)
	}
	return nil
}

type rawWrapper struct {
	APIVersion string          `json:"api_version"`
	Type       string          `json:"type"`
	ObjectMeta json.RawMessage `json:"metadata"`
	Spec       json.RawMessage `json:"spec"`
}
