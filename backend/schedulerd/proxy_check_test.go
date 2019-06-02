package schedulerd

import (
	"reflect"
	"testing"

	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
			entityAttributes: []string{`entity.name == "entity1"`},
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
			entityAttributes: []string{`entity.deregister == false`},
			entities: []*types.Entity{
				&types.Entity{Deregister: false},
				&types.Entity{Deregister: true},
			},
			want: []*types.Entity{
				&types.Entity{Deregister: false},
			},
		},
		{
			name:             "nested standard attribute",
			entityAttributes: []string{`entity.system.hostname == "foo.local"`},
			entities: []*types.Entity{
				&types.Entity{System: types.System{Hostname: "localhost"}},
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "foo"}},
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "foo"}, System: types.System{Hostname: "foo.local"}},
			},
			want: []*types.Entity{
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "foo"}, System: types.System{Hostname: "foo.local"}},
			},
		},
		{
			name:             "multiple matches",
			entityAttributes: []string{`entity.entity_class == "proxy"`},
			entities: []*types.Entity{
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "foo"}, EntityClass: "proxy"},
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "bar"}, EntityClass: "agent"},
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "baz"}, EntityClass: "proxy"},
			},
			want: []*types.Entity{
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "foo"}, EntityClass: "proxy"},
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "baz"}, EntityClass: "proxy"},
			},
		},
		{
			name:             "invalid expression",
			entityAttributes: []string{`foo &&`},
			entities: []*types.Entity{
				&types.Entity{ObjectMeta: types.ObjectMeta{Name: "foo"}},
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
	assert := assert.New(t)

	check := types.FixtureCheckConfig("check1")
	check.ProxyRequests = types.FixtureProxyRequests(true)

	// 10s * 90% / 3 = 3
	check.Interval = 10
	splay, err := calculateSplayInterval(check, 3)
	assert.Equal(3*time.Second, splay)
	assert.Nil(err)

	// 20s * 50% / 5 = 2
	check.Interval = 20
	check.ProxyRequests.SplayCoverage = 50
	splay, err = calculateSplayInterval(check, 5)
	assert.Equal(2*time.Second, splay)
	assert.Nil(err)

	// invalid cron string
	check.Cron = "invalid"
	splay, err = calculateSplayInterval(check, 5)
	assert.Equal(time.Duration(0), splay)
	assert.NotNil(err)

	// at most, 60s from current time * 50% / 2 = 15
	// this test will depend on when it is run, but the
	// largest splay calculation will be 15
	check.Cron = "* * * * *"
	splay, err = calculateSplayInterval(check, 2)
	assert.True(splay >= 0 && splay <= 15*time.Second)
	assert.Nil(err)
}

func TestSubstituteProxyEntityTokens(t *testing.T) {
	assert := assert.New(t)

	entity := types.FixtureEntity("entity1")
	check := types.FixtureCheckConfig("check1")
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = types.FixtureProxyRequests(true)

	substitutedProxyEntityTokens, err := substituteProxyEntityTokens(entity, check)
	if err != nil {
		assert.FailNow(err.Error())
	}
	assert.Equal(entity.Name, substitutedProxyEntityTokens.ProxyEntityName)
}

func BenchmarkMatchEntities1000(b *testing.B) {
	entity := corev2.FixtureEntity("foo")
	// non-matching expression to avoid short-circuiting behaviour
	expression := "entity.system.arch == 'amd65'"

	entities := make([]*corev2.Entity, 100)
	expressions := make([]string, 10)

	for i := range entities {
		entities[i] = entity
	}
	for i := range expressions {
		expressions[i] = expression
	}

	req := &corev2.ProxyRequests{EntityAttributes: expressions}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = matchEntities(entities, req)
	}
}
