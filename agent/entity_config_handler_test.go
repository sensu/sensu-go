package agent

import (
	"context"
	"testing"

	"github.com/gogo/protobuf/proto"
	corev3 "github.com/sensu/core/v3"
)

func TestHandleEntityConfig(t *testing.T) {
	cfg, cleanup := FixtureConfig()
	defer cleanup()
	a, err := NewAgent(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ecfg := corev3.FixtureEntityConfig("localhost.localdomain")
	state := a.getEntityState()
	ecfg.Metadata.Name = state.Metadata.Name
	b, err := a.marshal(ecfg)
	if err != nil {
		t.Fatal(err)
	}
	exp, err := corev3.V3EntityToV2(ecfg, state)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		<-a.entityConfigCh
	}()
	if err := a.handleEntityConfig(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	if got, want := a.getAgentEntity(), exp; !proto.Equal(got, want) {
		t.Errorf("bad entity; got %v, want %v", got, want)
	}
	go func() {
		<-a.entityConfigCh
	}()
	// this will cause an error, the state name will not match the cfg name
	ecfg.Metadata.Name = "foo"
	b, err = a.marshal(ecfg)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.handleEntityConfig(context.Background(), b); err != nil {
		t.Fatal(err)
	}
	if got, want := a.getAgentEntity(), exp; proto.Equal(got, want) {
		t.Error("expected returned entity to differ")
	}
}
