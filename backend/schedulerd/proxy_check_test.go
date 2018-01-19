package schedulerd

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/types"
)

func TestMatchEntities(t *testing.T) {
	tests := []struct {
		name             string
		entityAttributes []string
		entities         []*types.Entity
		want             []*types.Entity
	}{
		{
			name:             "standard string attribute",
			entityAttributes: []string{`entity.ID == "entity1"`},
			entities: []*types.Entity{
				types.FixtureEntity("entity1"),
				types.FixtureEntity("entity2"),
			},
			want: []*types.Entity{
				types.FixtureEntity("entity1"),
			},
		},
		{
			name:             "standard bool attribute",
			entityAttributes: []string{`entity.Deregister == false`},
			entities: []*types.Entity{
				&types.Entity{Deregister: false},
				&types.Entity{Deregister: true},
			},
			want: []*types.Entity{
				&types.Entity{Deregister: false},
			},
		},
		{
			name:             "extended attribute",
			entityAttributes: []string{`entity.Team == "dev"`},
			entities: []*types.Entity{
				&types.Entity{ExtendedAttributes: []byte(`{"Team": "dev"}`)},
				&types.Entity{ExtendedAttributes: []byte(`{"Team": "ops"}`)},
			},
			want: []*types.Entity{
				&types.Entity{ExtendedAttributes: []byte(`{"Team": "dev"}`)},
			},
		},
		{
			name:             "standard & extended attribute",
			entityAttributes: []string{`entity.Deregister == false`, `entity.Team == "dev"`},
			entities: []*types.Entity{
				&types.Entity{Deregister: false, ExtendedAttributes: []byte(`{"Team": "dev"}`)},
				&types.Entity{Deregister: true, ExtendedAttributes: []byte(`{"Team": "dev"}`)},
				&types.Entity{Deregister: false, ExtendedAttributes: []byte(`{"Team": "ops"}`)},
			},
			want: []*types.Entity{
				&types.Entity{Deregister: false, ExtendedAttributes: []byte(`{"Team": "dev"}`)},
			},
		},
		{
			name:             "nested standard attribute",
			entityAttributes: []string{`entity.System.Hostname == "foo.local"`},
			entities: []*types.Entity{
				&types.Entity{System: types.System{Hostname: "localhost"}},
				&types.Entity{ID: "foo"},
				&types.Entity{ID: "foo", System: types.System{Hostname: "foo.local"}},
			},
			want: []*types.Entity{
				&types.Entity{ID: "foo", System: types.System{Hostname: "foo.local"}},
			},
		},
		{
			name:             "nested extended attribute",
			entityAttributes: []string{`entity.Teams.Support == "dev"`},
			entities: []*types.Entity{
				&types.Entity{ExtendedAttributes: []byte(`{"Teams": {"Support": "dev"}}`)},
				&types.Entity{ExtendedAttributes: []byte(`{"Teams": {"Support": "ops"}}`)},
			},
			want: []*types.Entity{
				&types.Entity{ExtendedAttributes: []byte(`{"Teams": {"Support": "dev"}}`)},
			},
		},
		{
			name:             "invalid expression",
			entityAttributes: []string{`foo &&`},
			entities: []*types.Entity{
				&types.Entity{ID: "foo"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &types.ProxyRequests{
				EntityAttributes: tc.entityAttributes,
			}
			got := matchEntities(tc.entities, p)

			if len(got) != len(tc.want) {
				t.Errorf("Expected %d entities, got %d", len(tc.want), len(got))
				return
			}

			for i, _ := range tc.want {
				if !reflect.DeepEqual(got[i], tc.want[i]) {
					t.Errorf("MatchEntities() = %v, want %v", got, tc.want)
					return
				}
			}

		})
	}
}
