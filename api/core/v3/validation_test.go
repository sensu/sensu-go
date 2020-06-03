package v3

import (
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func TestValidateMetadata(t *testing.T) {
	if err := ValidateMetadata(nil); err == nil {
		t.Error("expected non-nil error")
	}

	var meta corev2.ObjectMeta

	if err := ValidateMetadata(&meta); err == nil {
		t.Error("expected non-nil error")
	}

	meta.Name = "foo"

	if err := ValidateMetadata(&meta); err == nil {
		t.Error("expected non-nil error")
	}

	meta.Labels = make(map[string]string)

	if err := ValidateMetadata(&meta); err == nil {
		t.Error("expected non-nil error")
	}

	meta.Annotations = make(map[string]string)

	if err := ValidateMetadata(&meta); err != nil {
		t.Error(err)
	}
}
