package postgres

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
)

func TestEntityConfigWrapperSQLParams(t *testing.T) {
	want := len(new(EntityConfigWrapper).SQLParams())
	got := reflect.ValueOf(EntityConfigWrapper{}).NumField()
	if got > want {
		t.Errorf("field added to EntityConfigWrapper but not SQLParams: got %d, want %d", got, want)
	}
	if got < want {
		t.Errorf("field removed from EntityConfigWrapper, but not SQLParams: got %d, want %d", got, want)
	}
}

func TestEntityConfigWrapperUnwrap(t *testing.T) {
	wrapper := EntityConfigWrapper{
		Namespace:   "default",
		Name:        "name",
		Selectors:   []byte(`{"labels.foo":"bar"}`),
		Annotations: []byte(`{"anno":"t8n"}`),
		EntityClass: "agent",
	}
	resource, err := wrapper.Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	cfg := resource.(*corev3.EntityConfig)
	if err := cfg.Validate(); err != nil {
		t.Fatal(err)
	}
	wantMeta := &corev2.ObjectMeta{
		Namespace:   "default",
		Name:        "name",
		Labels:      map[string]string{"foo": "bar"},
		Annotations: map[string]string{"anno": "t8n"},
	}
	if got, want := cfg.Metadata, wantMeta; !reflect.DeepEqual(got, want) {
		t.Errorf("bad metadata: got %v, want %v", got, want)
	}
	if got, want := cfg.EntityClass, wrapper.EntityClass; got != want {
		t.Errorf("bad entity_class: got %v, want %v", got, want)
	}
	if got, want := cfg.User, wrapper.User; got != want {
		t.Errorf("bad user: got %v, want %v", got, want)
	}
}

func TestEntityConfigWrapperWrapUnwrap(t *testing.T) {
	cfg := corev3.FixtureEntityConfig("testent")
	got := WrapEntityConfig(cfg)
	resource, err := got.Unwrap()
	if err != nil {
		t.Fatal(err)
	}
	if err := resource.Validate(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cfg, resource) {
		t.Errorf("wrap/unwrap cycle yielded different results: got %#v, want %#v", resource.GetMetadata(), cfg.GetMetadata())
	}
}
