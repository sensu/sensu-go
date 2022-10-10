package globalid

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/testing/fixture"
)

// DefaultEncoder is the default implementation of Encode.
func TestDefaultEncoder(t *testing.T) {
	tc := []struct {
		name     string
		in       interface{}
		expectId string
		expectNs string
	}{
		{
			name: "v2 resource",
			in: &fixture.Resource{
				ObjectMeta: corev2.ObjectMeta{
					Name:      "x",
					Namespace: "y",
				},
			},
			expectId: "x",
			expectNs: "y",
		},
		{
			name: "v3 resource",
			in: &fixture.V3Resource{
				Metadata: &corev2.ObjectMeta{
					Name:      "x",
					Namespace: "y",
				},
			},
			expectId: "x",
			expectNs: "y",
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			cmp := DefaultEncoder(context.Background(), tt.in)
			if cmp.Namespace() != tt.expectNs {
				t.Errorf("DefaultEncoder() expect = %s, got = %s", cmp.Namespace(), tt.expectNs)
			}
			if cmp.UniqueComponent() != tt.expectId {
				t.Errorf("DefaultEncoder() expect = %s, got = %s", cmp.UniqueComponent(), tt.expectId)
			}
		})
	}
}
