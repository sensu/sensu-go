package graphql

import (
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalFilters(t *testing.T) {
	fs := DefaultGlobalFilters()
	require.NotEmpty(t, fs)

	testCases := []struct {
		statement   string
		setupRecord func() corev3.Resource
		fieldsFunc  filter.FieldsFunc
		expect      bool
	}{
		{
			statement: `labelSelector:region == "us-west-2"`,
			expect:    true,
			setupRecord: func() corev3.Resource {
				ent := corev2.FixtureEntity("a")
				ent.Labels = map[string]string{"region": "us-west-2"}
				return ent
			},
			fieldsFunc: corev3.EntityFields,
		},
		{
			statement: `labelSelector:region == "us-east-1"`,
			expect:    false,
			setupRecord: func() corev3.Resource {
				ent := corev2.FixtureEntity("a")
				ent.Labels = map[string]string{"region": "us-west-2"}
				return ent
			},
			fieldsFunc: corev3.EntityFields,
		},
		{
			statement: `labelSelector:entity.region is "us-west-2"`,
			expect:    false,
			setupRecord: func() corev3.Resource {
				ent := corev2.FixtureEntity("a")
				ent.Labels = map[string]string{"region": "us-west-2"}
				return ent
			},
			fieldsFunc: corev3.EntityFields,
		},
		{
			statement: `fieldSelector:entity.entity_class == "proxy"`,
			expect:    true,
			setupRecord: func() corev3.Resource {
				ent := corev2.FixtureEntity("a")
				ent.EntityClass = "proxy"
				return ent
			},
			fieldsFunc: corev3.EntityFields,
		},
		{
			statement: `fieldSelector:entity.entityClass == "foo"`,
			expect:    false,
			setupRecord: func() corev3.Resource {
				ent := corev2.FixtureEntity("a")
				ent.EntityClass = "bar"
				return ent
			},
			fieldsFunc: corev3.EntityFields,
		},
		{
			statement: `fieldSelector:entity.entityClass is "foo"`,
			expect:    false,
			setupRecord: func() corev3.Resource {
				ent := corev2.FixtureEntity("a")
				ent.EntityClass = "bar"
				return ent
			},
			fieldsFunc: corev3.EntityFields,
		},
		{
			statement: `labelSelector:region == "na"`,
			expect:    true,
			setupRecord: func() corev3.Resource {
				return &corev2.Event{
					ObjectMeta: corev2.ObjectMeta{},
					Check: &corev2.Check{
						ObjectMeta: corev2.ObjectMeta{
							Labels: map[string]string{"region": "na"},
						},
					},
					Entity: &corev2.Entity{
						ObjectMeta: corev2.ObjectMeta{
							Labels: map[string]string{"country": "canada"},
						},
					},
				}
			},
			fieldsFunc: corev3.EventFields,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.statement, func(t *testing.T) {
			matches, err := filter.Compile([]string{tc.statement}, fs, tc.fieldsFunc)
			require.NoError(t, err)

			record := tc.setupRecord()
			assert.Equal(t, tc.expect, matches(record))
		})
	}
}
