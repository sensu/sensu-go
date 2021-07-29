package v3

// automatically generated file, do not edit!

import (
	"reflect"
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

func TestResolveV2ResourceEntityConfig(t *testing.T) {
	v2Resource, err := ResolveV2Resource("EntityConfig")
	if err != nil {
		t.Fatal(err)
	}
	v3Resource, err := ResolveResource("EntityConfig")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v2Resource.(*V2ResourceProxy).Resource, v3Resource; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad resource: got %v, want %v", got, want)
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

func TestResolveV2ResourceEntityState(t *testing.T) {
	v2Resource, err := ResolveV2Resource("EntityState")
	if err != nil {
		t.Fatal(err)
	}
	v3Resource, err := ResolveResource("EntityState")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v2Resource.(*V2ResourceProxy).Resource, v3Resource; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad resource: got %v, want %v", got, want)
	}
}

func TestResolvePipeline(t *testing.T) {
	var value interface{} = new(Pipeline)
	if _, ok := value.(Resource); ok {
		resource, err := ResolveResource("Pipeline")
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
	_, err := ResolveResource("Pipeline")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"Pipeline" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolvePipelineByRBACName(t *testing.T) {
	value := new(Pipeline)
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

func TestResolvePipelineByStoreName(t *testing.T) {
	value := new(Pipeline)
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

func TestResolveV2ResourcePipeline(t *testing.T) {
	v2Resource, err := ResolveV2Resource("Pipeline")
	if err != nil {
		t.Fatal(err)
	}
	v3Resource, err := ResolveResource("Pipeline")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v2Resource.(*V2ResourceProxy).Resource, v3Resource; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad resource: got %v, want %v", got, want)
	}
}

func TestResolveResourceTemplate(t *testing.T) {
	var value interface{} = new(ResourceTemplate)
	if _, ok := value.(Resource); ok {
		resource, err := ResolveResource("ResourceTemplate")
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
	_, err := ResolveResource("ResourceTemplate")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if got, want := err.Error(), `"ResourceTemplate" is not a Resource`; got != want {
		t.Fatalf("unexpected error: %s", err)
	}
}

func TestResolveResourceTemplateByRBACName(t *testing.T) {
	value := new(ResourceTemplate)
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

func TestResolveResourceTemplateByStoreName(t *testing.T) {
	value := new(ResourceTemplate)
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

func TestResolveV2ResourceResourceTemplate(t *testing.T) {
	v2Resource, err := ResolveV2Resource("ResourceTemplate")
	if err != nil {
		t.Fatal(err)
	}
	v3Resource, err := ResolveResource("ResourceTemplate")
	if err != nil {
		t.Fatal(err)
	}
	if got, want := v2Resource.(*V2ResourceProxy).Resource, v3Resource; !reflect.DeepEqual(got, want) {
		t.Fatalf("bad resource: got %v, want %v", got, want)
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
