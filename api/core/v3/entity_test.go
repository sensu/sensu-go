package v3

import (
	"reflect"
	"testing"

	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestV2EntityToV3(t *testing.T) {
	entity := corev2.FixtureEntity("foo")
	cfg, state := V2EntityToV3(entity)
	if got, want := cfg.Metadata, entity.ObjectMeta; !proto.Equal(got, &want) {
		t.Errorf("bad objectmeta: got %v, want %v", got, want)
	}
	if got, want := cfg.Metadata.Labels, make(map[string]string); !reflect.DeepEqual(got, want) {
		t.Errorf("bad objectmeta labels: got %v, want %v", got, want)
	}
	if got, want := cfg.Metadata.Annotations, make(map[string]string); !reflect.DeepEqual(got, want) {
		t.Errorf("bad objectmeta annotations: got %v, want %v", got, want)
	}
	if got, want := cfg.EntityClass, entity.EntityClass; got != want {
		t.Errorf("bad EntityClass: got %v, want %v", got, want)
	}
	if got, want := cfg.User, entity.User; got != want {
		t.Errorf("bad User: got %v, want %v", got, want)
	}
	if got, want := cfg.Subscriptions, entity.Subscriptions; !reflect.DeepEqual(got, want) {
		t.Errorf("bad Subscriptions: got %v, want %v", got, want)
	}
	if got, want := cfg.Deregister, entity.Deregister; got != want {
		t.Errorf("bad Deregister: got %v, want %v", got, want)
	}
	if got, want := cfg.Deregistration, entity.Deregistration; !proto.Equal(&got, &want) {
		t.Errorf("bad Deregistration: got %v, want %v", got, want)
	}
	if got, want := cfg.KeepaliveHandlers, entity.KeepaliveHandlers; !reflect.DeepEqual(got, want) {
		t.Errorf("bad KeepaliveHandlers: got %v, want %v", got, want)
	}
	if got, want := cfg.Redact, entity.Redact; !reflect.DeepEqual(got, want) {
		t.Errorf("bad Redact: got %v, want %v", got, want)
	}
	if got, want := state.Metadata, entity.ObjectMeta; !proto.Equal(got, &want) {
		t.Errorf("bad objectmeta: got %v, want %v", got, want)
	}
	if got, want := state.Metadata.Labels, make(map[string]string); !reflect.DeepEqual(got, want) {
		t.Errorf("bad objectmeta labels: got %v, want %v", got, want)
	}
	if got, want := state.Metadata.Annotations, make(map[string]string); !reflect.DeepEqual(got, want) {
		t.Errorf("bad objectmeta annotations: got %v, want %v", got, want)
	}
	if got, want := state.System, entity.System; !proto.Equal(&got, &want) {
		t.Errorf("bad System: got %v, want %v", got, want)
	}
	if got, want := state.LastSeen, entity.LastSeen; got != want {
		t.Errorf("bad LastSeen: got %v, want %v", got, want)
	}
	if got, want := state.SensuAgentVersion, entity.SensuAgentVersion; got != want {
		t.Errorf("bad SensuAgentVersion: got %v, want %v", got, want)
	}
}

func TestV3EntityToV2(t *testing.T) {
	cfg := FixtureEntityConfig("foo")
	state := FixtureEntityState("foo")
	entity, err := V3EntityToV2(cfg, state)
	if err != nil {
		t.Fatal(err)
	}

	if want, got := cfg.Metadata, entity.ObjectMeta; !proto.Equal(&got, want) {
		t.Errorf("bad objectmeta: got %v, want %v", got, want)
	}
	if want, got := cfg.EntityClass, entity.EntityClass; got != want {
		t.Errorf("bad EntityClass: got %v, want %v", got, want)
	}
	if want, got := cfg.User, entity.User; got != want {
		t.Errorf("bad User: got %v, want %v", got, want)
	}
	if want, got := cfg.Subscriptions, entity.Subscriptions; !reflect.DeepEqual(got, want) {
		t.Errorf("bad Subscriptions: got %v, want %v", got, want)
	}
	if want, got := cfg.Deregister, entity.Deregister; got != want {
		t.Errorf("bad Deregister: got %v, want %v", got, want)
	}
	if want, got := cfg.Deregistration, entity.Deregistration; !proto.Equal(&got, &want) {
		t.Errorf("bad Deregistration: got %v, want %v", got, want)
	}
	if want, got := cfg.KeepaliveHandlers, entity.KeepaliveHandlers; !reflect.DeepEqual(got, want) {
		t.Errorf("bad KeepaliveHandlers: got %v, want %v", got, want)
	}
	if want, got := cfg.Redact, entity.Redact; !reflect.DeepEqual(got, want) {
		t.Errorf("bad Redact: got %v, want %v", got, want)
	}
	if want, got := state.System, entity.System; !proto.Equal(&got, &want) {
		t.Errorf("bad System: got %v, want %v", got, want)
	}
	if want, got := state.LastSeen, entity.LastSeen; got != want {
		t.Errorf("bad LastSeen: got %v, want %v", got, want)
	}
	if want, got := state.SensuAgentVersion, entity.SensuAgentVersion; got != want {
		t.Errorf("bad SensuAgentVersion: got %v, want %v", got, want)
	}

	var blankC EntityConfig
	if _, err := V3EntityToV2(&blankC, state); err == nil {
		t.Errorf("expected non-nil error")
	}
	var blankS EntityState
	if _, err := V3EntityToV2(cfg, &blankS); err == nil {
		t.Errorf("expected non-nil error")
	}

	blankC.Metadata = &corev2.ObjectMeta{
		Namespace: "default",
		Name:      "foo",
	}

	blankS.Metadata = &corev2.ObjectMeta{
		Namespace: "notdefault",
		Name:      "foo",
	}

	if _, err := V3EntityToV2(&blankC, &blankS); err == nil {
		t.Errorf("expected non-nil error")
	}

	blankS.Metadata.Namespace = "default"
	blankS.Metadata.Name = "notfoo"

	if _, err := V3EntityToV2(&blankC, &blankS); err == nil {
		t.Errorf("expected non-nil error")
	}
}

func TestFixtureEntityConfig(t *testing.T) {
	if err := FixtureEntityConfig("foo").Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestFixtureEntityState(t *testing.T) {
	if err := FixtureEntityState("foo").Validate(); err != nil {
		t.Fatal(err)
	}
}
