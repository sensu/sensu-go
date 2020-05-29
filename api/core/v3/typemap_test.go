package v3

// automatically generated file, do not edit!

import (
	"testing"
)

func TestResolveEntityConfig(t *testing.T) {
	var value interface{} = new(EntityConfig)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("EntityConfig"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("EntityConfig")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"EntityConfig" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveEntityState(t *testing.T) {
	var value interface{} = new(EntityState)
	if _, ok := value.(Resource); ok {
		if _, err := ResolveResource("EntityState"); err != nil {
			t.Fatal(err)
		}
		return
	}
	_, err := ResolveResource("EntityState")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"EntityState" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveNotExists(t *testing.T) {
	_, err := ResolveResource("!#$@$%@#$")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}
