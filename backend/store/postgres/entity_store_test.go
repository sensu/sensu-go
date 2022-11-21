package postgres

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/types"
)

var wrapper = NewResourceWrapper(storev2.WrapResource)

func init() {
	// This causes storev2.WrapResource to use postgres. Outside of test contexts,
	// this is does in the store provider, which knows when postgres is being enabled
	// or disabled.
	wrapper.EnablePostgres()
	storev2.WrapResource = wrapper.WrapResource
}

func TestEntityStorage(t *testing.T) {
	testWithPostgresStore(t, func(str storev2.Interface) {
		db := str.(*Store).db
		s := NewEntityStore(db)
		entity := corev2.FixtureEntity("entity")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, entity.Namespace)
		pred := &store.SelectionPredicate{}

		// We should receive an empty slice if no results were found
		entities, err := s.GetEntities(ctx, pred)
		if err != nil {
			t.Fatal(err)
		}
		if entities == nil {
			t.Fatal("nil entities")
		}
		if got, want := pred.Continue, ""; got != want {
			t.Errorf("bad pred.Continue: got %q, want %q", got, want)
		}

		namespace := corev3.FixtureNamespace(entity.Namespace)
		if err := str.GetNamespaceStore().CreateOrUpdate(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}

		retrieved, err := s.GetEntityByName(ctx, entity.Name)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := retrieved.Name, entity.Name; got != want {
			t.Errorf("bad name: got %q, want %q", got, want)
		}

		entities, err = s.GetEntities(ctx, pred)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := len(entities), 1; got != want {
			t.Fatalf("wrong number of entities: got %d, want %d", got, want)
		}
		if got, want := entities[0].Name, entity.Name; got != want {
			t.Errorf("bad entity name: got %q, want %q", got, want)
		}
		if got, want := pred.Continue, ""; got != want {
			t.Errorf("bad pred.Continue: got %q, want %q", got, want)
		}

		if err := s.DeleteEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}

		retrieved, err = s.GetEntityByName(ctx, entity.Name)
		if err != nil {
			t.Fatal(err)
		}
		if retrieved != nil {
			t.Fatalf("want nil, got %v", retrieved)
		}

		// Nonexistent entity deletion should return no error.
		if err := s.DeleteEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}

		// Updating an entity in a nonexistent namespace should not work
		entity.Namespace = "missing"
		if err = s.UpdateEntity(ctx, entity); err == nil {
			t.Errorf("expected non-nil error")
		}
	})
}

func TestEntityIteration(t *testing.T) {
	configs := []*corev3.EntityConfig{
		corev3.FixtureEntityConfig("a"),
		corev3.FixtureEntityConfig("b"),
		corev3.FixtureEntityConfig("c"),
		corev3.FixtureEntityConfig("d"),
		corev3.FixtureEntityConfig("e"),
	}
	states := map[uniqueResource]*corev3.EntityState{
		uniqueResource{Name: "b", Namespace: "default"}: corev3.FixtureEntityState("b"),
		uniqueResource{Name: "c", Namespace: "default"}: corev3.FixtureEntityState("c"),
		uniqueResource{Name: "d", Namespace: "default"}: corev3.FixtureEntityState("d"),
	}
	entities, err := entitiesFromConfigsAndStates(configs, states)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(entities), len(configs); got != want {
		t.Fatalf("bad entity count: got %d, want %d", got, want)
	}
	for _, config := range configs {
		match := false
		for _, entity := range entities {
			if config.Metadata.Name == entity.ObjectMeta.Name {
				match = true
				break
			}
		}
		if !match {
			t.Errorf("entity was not found with name: %q", config.Metadata.Name)
		}
	}
}

func TestEntityIterationNoPanicMismatched(t *testing.T) {
	configs := []*corev3.EntityConfig{
		corev3.FixtureEntityConfig("b"),
		corev3.FixtureEntityConfig("c"),
	}
	states := map[uniqueResource]*corev3.EntityState{
		{Name: "a", Namespace: "default"}: corev3.FixtureEntityState("a"),
		{Name: "b", Namespace: "default"}: corev3.FixtureEntityState("b"),
		{Name: "c", Namespace: "default"}: corev3.FixtureEntityState("c"),
	}
	if _, err := entitiesFromConfigsAndStates(configs, states); err != nil {
		t.Fatal(err)
	}
}

func TestEntityCreateOrUpdateMultipleAddresses(t *testing.T) {
	testWithPostgresStore(t, func(str storev2.Interface) {
		db := str.(*Store).db
		s := NewEntityStore(db)
		entity := types.FixtureEntity("entity")
		ctx := context.WithValue(context.Background(), types.NamespaceKey, entity.Namespace)
		entity.System.Network = corev2.Network{
			Interfaces: []corev2.NetworkInterface{
				{
					Name:      "a",
					MAC:       "asdfasdfasdf",
					Addresses: []string{"127.0.0.1/16", "::1/128"},
				},
				{
					Name:      "b",
					MAC:       "adlfidfasdfasdf",
					Addresses: []string{"127.0.0.1/8", "::1/128"},
				},
			},
		}
		entity.System.Network.Interfaces[0].Addresses = append(entity.System.Network.Interfaces[0].Addresses, "1.1.1.1")
		namespace := corev3.FixtureNamespace(entity.Namespace)
		if err := str.GetNamespaceStore().CreateOrUpdate(ctx, namespace); err != nil {
			t.Fatal(err)
		}
		if err := s.UpdateEntity(ctx, entity); err != nil {
			t.Fatal(err)
		}
	})
}
