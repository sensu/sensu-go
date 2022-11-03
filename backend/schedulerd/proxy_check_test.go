package schedulerd

import (
	"reflect"
	"testing"

	time "github.com/echlebek/timeproxy"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	"github.com/stretchr/testify/assert"
)

func TestMatchEntities(t *testing.T) {
	entity1 := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:      "entity1",
			Namespace: "default",
			Labels:    map[string]string{"proxy_type": "switch"},
		},
		EntityClass: "proxy",
	}
	entity2 := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:      "entity2",
			Namespace: "default",
			Labels:    map[string]string{"proxy_type": "sensor"},
		},
		Deregister:  true,
		EntityClass: "proxy",
	}
	entity3 := &corev3.EntityConfig{
		Metadata: &corev2.ObjectMeta{
			Name:      "entity3",
			Namespace: "default",
		},
		EntityClass: "agent",
	}

	tests := []struct {
		name             string
		entityAttributes []string
		entities         []corev3.Resource
		want             []*corev3.EntityConfig
	}{
		{
			name:             "standard string attribute",
			entityAttributes: []string{`entity.name == "entity1"`},
			entities:         []corev3.Resource{entity1, entity2, entity3},
			want:             []*corev3.EntityConfig{entity1},
		},
		{
			name:             "standard bool attribute",
			entityAttributes: []string{`entity.deregister == true`},
			entities:         []corev3.Resource{entity1, entity2, entity3},
			want:             []*corev3.EntityConfig{entity2},
		},
		{
			name:             "nested standard attribute",
			entityAttributes: []string{`entity.metadata.name == "entity1"`},
			entities:         []corev3.Resource{entity1, entity2, entity3},
			want:             []*corev3.EntityConfig{entity1},
		},
		{
			name:             "multiple matches",
			entityAttributes: []string{`entity.entity_class == "proxy"`},
			entities:         []corev3.Resource{entity1, entity2, entity3},
			want:             []*corev3.EntityConfig{entity1, entity2},
		},
		{
			name:             "invalid expression",
			entityAttributes: []string{`foo &&`},
			entities:         []corev3.Resource{entity1, entity2, entity3},
		},
		{
			name: "multiple entity attributes",
			entityAttributes: []string{
				`entity.entity_class == "proxy"`,
				`entity.metadata.labels.proxy_type == "sensor"`,
			},
			entities: []corev3.Resource{entity1, entity2, entity3},
			want:     []*corev3.EntityConfig{entity2},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := &corev2.ProxyRequests{
				EntityAttributes: tc.entityAttributes,
			}
			cacher := cachev2.NewFromResources(tc.entities, true)
			got := matchEntities(cacher.Get("default"), p)

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

	check := corev2.FixtureCheckConfig("check1")
	check.ProxyRequests = corev2.FixtureProxyRequests(true)

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

	entity := corev3.FixtureEntityConfig("entity1")
	check := corev2.FixtureCheckConfig("check1")
	check.Subscriptions = []string{"subscription1"}
	check.ProxyRequests = corev2.FixtureProxyRequests(true)

	substitutedProxyEntityTokens, err := substituteProxyEntityTokens(entity, check)
	if err != nil {
		assert.FailNow(err.Error())
	}
	assert.Equal(entity.Metadata.Name, substitutedProxyEntityTokens.ProxyEntityName)
}

func BenchmarkMatchEntities1000(b *testing.B) {
	entity := corev3.FixtureEntityConfig("foo")
	// non-matching expression to avoid short-circuiting behaviour
	expression := "entity.system.arch == 'amd65'"

	entities := make([]corev3.Resource, 100)
	expressions := make([]string, 10)

	for i := range entities {
		entities[i] = entity
	}
	for i := range expressions {
		expressions[i] = expression
	}

	req := &corev2.ProxyRequests{EntityAttributes: expressions}
	cacher := cachev2.NewFromResources(entities, true)
	resources := cacher.Get("default")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = matchEntities(resources, req)
	}
}
