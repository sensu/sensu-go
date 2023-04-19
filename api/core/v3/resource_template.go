package v3

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	types "github.com/sensu/sensu-go/types"
)

// ResourceTemplate is a template for core/v3 resources.
type ResourceTemplate struct {
	Metadata   *corev2.ObjectMeta `json:"metadata"`
	APIVersion string             `json:"api_version"`
	Type       string             `json:"type"`
	Template   string             `json:"template"`
}

func (r *ResourceTemplate) GetMetadata() *corev2.ObjectMeta {
	return r.Metadata
}

// Execute executes the Template in the ResourceTemplate. It's given a metadata
// object to draw variables from. Typically, templating will be done with a
// variable namespace or name, but could also be done with labels and annotations
// from the metadata, depending on the nature of the template.
func (r *ResourceTemplate) Execute(meta *corev2.ObjectMeta) (Resource, error) {
	tmpl, err := template.New("resource").Parse(r.Template)
	if err != nil {
		return nil, fmt.Errorf("error parsing resource template: %s", err)
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, meta); err != nil {
		return nil, fmt.Errorf("error executing resource template: %s", err)
	}
	t, err := types.ResolveRaw(r.APIVersion, r.Type)
	if err != nil {
		// do not wrap this error
		return nil, err
	}
	resource, ok := t.(Resource)
	if !ok {
		return nil, fmt.Errorf("error expanding resource template: %T is not a core/v3 resource", t)
	}
	if err := json.Unmarshal(buf.Bytes(), &resource); err != nil {
		return nil, fmt.Errorf("error expanding resource template: invalid json: %s", err)
	}
	if err := resource.Validate(); err != nil {
		return nil, fmt.Errorf("error expanding resource template: resource not valid: %s", err)
	}
	return resource, nil
}

func (r *ResourceTemplate) validate() error {
	if _, err := template.New("validate").Parse(r.Template); err != nil {
		return err
	}
	if _, err := types.ResolveRaw(r.APIVersion, r.Type); err != nil {
		return err
	}
	return nil
}
