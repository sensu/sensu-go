package v3_test

import (
	"reflect"
	testing "testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	v3 "github.com/sensu/sensu-go/api/core/v3"
)

func TestResourceTemplateExecute(t *testing.T) {
	template := &v3.ResourceTemplate{
		Metadata: &corev2.ObjectMeta{
			Name: "template",
		},
		APIVersion: "core/v3",
		Type:       "EntityConfig",
		Template: `
		{
			"metadata": {
				"namespace": "{{ .Namespace }}",
				"name": "entity1"
			},
			"entity_class": "agent",
			"user": "agent",
			"subscriptions": ["a", "b", "c"]
		}
		`,
	}

	metadata := &corev2.ObjectMeta{
		Namespace: "myns",
	}

	got, err := template.Execute(metadata)
	if err != nil {
		t.Fatal(err)
	}

	want := &v3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:        "entity1",
			Namespace:   "myns",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		EntityClass:   "agent",
		User:          "agent",
		Subscriptions: []string{"a", "b", "c"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("bad resource: got %#v, want %#v", got, want)
	}

	// Causes entityconfig validation to fail
	metadata.Namespace = ""
	if _, err := template.Execute(metadata); err == nil {
		t.Error("expected non-nil error")
	}
	metadata.Namespace = "myns"

	// causes type lookup to fail
	template.APIVersion = "notexists"
	if _, err := template.Execute(metadata); err == nil {
		t.Error("expected non-nil error")
	}
	template.APIVersion = "core/v3"

	// causes json unmarshaling to fail
	template.Template = `{sdlfkjsdlfkj}`
	if _, err := template.Execute(metadata); err == nil {
		t.Error("expected non-nil error")
	}
}

func TestResourceTemplateValidate(t *testing.T) {
	tmpl := &v3.ResourceTemplate{
		Metadata: &corev2.ObjectMeta{
			Name:        "foobar",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		APIVersion: "core/v3",
		Type:       "EntityConfig",
		Template:   "{{ .Foobar }}",
	}
	if err := tmpl.Validate(); err != nil {
		t.Error(err)
	}
	tmpl.Template = "{{"
	if err := tmpl.Validate(); err == nil {
		t.Error("expected non-nil error")
	}
	tmpl.Template = "{{ .Foobar }}"
	tmpl.APIVersion = "fake/v2"
	if err := tmpl.Validate(); err == nil {
		t.Error("expected non-nil error")
	}
}
