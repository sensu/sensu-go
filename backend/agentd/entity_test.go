package agentd

import (
	"reflect"
	"testing"
)

func TestAddEntitySubscription(t *testing.T) {
	tests := []struct {
		name          string
		entityName    string
		subscriptions []string
		want          []string
	}{
		{
			name:          "the entity subscription is added if missing",
			entityName:    "foo",
			subscriptions: []string{},
			want:          []string{"entity:foo"},
		},
		{
			name:          "the entity subscription is not added if already present",
			entityName:    "foo",
			subscriptions: []string{"entity:foo"},
			want:          []string{"entity:foo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := addEntitySubscription(tt.entityName, tt.subscriptions); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("addEntitySubscription() = %v, want %v", got, tt.want)
			}
		})
	}
}
