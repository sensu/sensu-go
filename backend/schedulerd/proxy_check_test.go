package schedulerd

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
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
			name:             "multiple matches",
			entityAttributes: []string{`entity.Class == "proxy"`},
			entities: []*types.Entity{
				&types.Entity{ID: "foo", Class: "proxy"},
				&types.Entity{ID: "bar", Class: "agent"},
				&types.Entity{ID: "baz", Class: "proxy"},
			},
			want: []*types.Entity{
				&types.Entity{ID: "foo", Class: "proxy"},
				&types.Entity{ID: "baz", Class: "proxy"},
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

			for i := range tc.want {
				if !reflect.DeepEqual(got[i], tc.want[i]) {
					t.Errorf("MatchEntities() = %v, want %v", got, tc.want)
					return
				}
			}

		})
	}
}

func TestSplayCalculation(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	check := types.FixtureCheckConfig("check1")
	check.ProxyRequests = types.FixtureProxyRequests(true)

	// 10s * 90% / 3 = 3
	check.Interval = 10
	splay, err := calculateSplayInterval(check, 3)
	assert.Equal(float64(3), splay)
	assert.Nil(err)

	// 20s * 50% / 5 = 2
	check.Interval = 20
	check.ProxyRequests.SplayCoverage = 50
	splay, err = calculateSplayInterval(check, 5)
	assert.Equal(float64(2), splay)
	assert.Nil(err)

	// invalid cron string
	check.Cron = "invalid"
	splay, err = calculateSplayInterval(check, 5)
	assert.Equal(float64(0), splay)
	assert.NotNil(err)

	// at most, 60s from current time * 50% / 2 = 15
	// this test will depend on when it is run, but the
	// largest splay calculation will be 15
	check.Cron = "* * * * *"
	splay, err = calculateSplayInterval(check, 2)
	assert.True(splay >= 0 && splay <= 15)
	assert.Nil(err)
}

func TestSubstituteProxyEntityTokens(t *testing.T) {
	t.Parallel()

	assert := assert.New(t)

	entity := types.FixtureEntity("entity1")
	check := types.FixtureCheckConfig("check1")
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

	substitutedProxyEntityTokens, err := substituteProxyEntityTokens(entity, check)
	if err != nil {
		assert.FailNow(err.Error())
	}
	assert.Equal(entity.ID, substitutedProxyEntityTokens.ProxyEntityID)
}
