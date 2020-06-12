package v3

// automatically generated file, do not edit!

import (
	"testing"
)

func TestResolveEntityConfig(t *testing.T) {
	var value interface{} = new(EntityConfig)
	if _, ok := value.(Resource); ok {
		resource, err := ResolveResource("EntityConfig")
		if err != nil {
			t.Fatal(err)
		}
		meta := resource.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if meta.Labels == nil {
			t.Error("nil metadata")
		}
		if meta.Annotations == nil {
			t.Error("nil annotations")
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

func TestResolveEntityConfigByRBACName(t *testing.T) {
	value := new(EntityConfig)
	var iface interface{} = value
	resource, err := ResolveResourceByRBACName(value.RBACName())
	if _, ok := iface.(Resource); ok {
		if err != nil {
			t.Fatal(err)
		}
		meta := resource.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if meta.Labels == nil {
			t.Error("nil labels")
		}
		if meta.Annotations == nil {
			t.Errorf("nil annotations")
		}
	} else {
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	}
}

func TestResolveEntityConfigByStoreName(t *testing.T) {
	value := new(EntityConfig)
	var iface interface{} = value
	resource, err := ResolveResourceByStoreName(value.StoreName())
	if _, ok := iface.(Resource); ok {
		if err != nil {
			t.Fatal(err)
		}
		meta := resource.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if meta.Labels == nil {
			t.Error("nil labels")
		}
		if meta.Annotations == nil {
			t.Errorf("nil annotations")
		}
	} else {
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	}
}

func TestResolveEntityState(t *testing.T) {
	var value interface{} = new(EntityState)
	if _, ok := value.(Resource); ok {
		resource, err := ResolveResource("EntityState")
		if err != nil {
			t.Fatal(err)
		}
		meta := resource.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if meta.Labels == nil {
			t.Error("nil metadata")
		}
		if meta.Annotations == nil {
			t.Error("nil annotations")
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

func TestResolveEntityStateByRBACName(t *testing.T) {
	value := new(EntityState)
	var iface interface{} = value
	resource, err := ResolveResourceByRBACName(value.RBACName())
	if _, ok := iface.(Resource); ok {
		if err != nil {
			t.Fatal(err)
		}
		meta := resource.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if meta.Labels == nil {
			t.Error("nil labels")
		}
		if meta.Annotations == nil {
			t.Errorf("nil annotations")
		}
	} else {
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	}
}

func TestResolveEntityStateByStoreName(t *testing.T) {
	value := new(EntityState)
	var iface interface{} = value
	resource, err := ResolveResourceByStoreName(value.StoreName())
	if _, ok := iface.(Resource); ok {
		if err != nil {
			t.Fatal(err)
		}
		meta := resource.GetMetadata()
		if meta == nil {
			t.Fatal("nil metadata")
		}
		if meta.Labels == nil {
			t.Error("nil labels")
		}
		if meta.Annotations == nil {
			t.Errorf("nil annotations")
		}
	} else {
		if err == nil {
			t.Fatal("expected non-nil error")
		}
	}
}

func TestResolveNotExists(t *testing.T) {
	_, err := ResolveResource("!#$@$%@#$")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
}

func TestListResources(t *testing.T) {
	resources := ListResources()
	if got, want := len(resources), len(typeMap)/2; got != want {
		t.Fatalf("wrong number of resources: got %d, want %d", got, want)
	}
	for _, r := range resources {
		if r.GetMetadata() == nil {
			t.Errorf("nil metadata for resource %s", r.RBACName())
		}
	}
}
