package labels

import (
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/selector"
	"github.com/sensu/sensu-go/types"
	coreTypes "github.com/sensu/sensu-go/types"
)

type foo struct{}

type mockSelectorFunc = func(set map[string]string) bool

var mockSelectorReturn = func(set map[string]string) bool {
	if name, ok := set["return"]; ok && name == "true" {
		return true
	}
	return false
}

var mockRegionSelector = func(set map[string]string) bool {
	return set["country"] == "canada"
}

func TestFilter(t *testing.T) {
	resource1 := corev2.FixtureEntity("test1")
	resource1.ObjectMeta.Labels = map[string]string{"return": "true"}
	resource2 := corev2.FixtureEntity("test2")
	resource2.ObjectMeta.Labels = map[string]string{"return": "false"}
	resource3 := corev2.FixtureCheckConfig("test3")
	resource3W := types.WrapResource(resource3)
	resource3.ObjectMeta.Labels = map[string]string{"return": "true"}
	resource4 := corev2.FixtureCheckConfig("test4")
	resource4W := types.WrapResource(resource4)
	event := &corev2.Event{
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
	labelSelector, _ := selector.ParseLabelSelector("region == 'na' && country == 'canada'")

	tests := []struct {
		name             string
		resources        interface{}
		mockSelectorFunc mockSelectorFunc
		want             interface{}
	}{
		{
			name:             "resources",
			resources:        []corev2.Resource{resource1, resource2},
			mockSelectorFunc: mockSelectorReturn,
			want:             []corev2.Resource{resource1},
		},
		{
			name:             "wrapped resources",
			resources:        []*coreTypes.Wrapper{&resource3W, &resource4W},
			mockSelectorFunc: mockSelectorReturn,
			want:             []*coreTypes.Wrapper{&resource3W},
		},
		{
			name:             "unknown type",
			resources:        []*foo{},
			mockSelectorFunc: mockSelectorReturn,
			want:             []*foo{},
		},
		{
			name:             "event's nested check label",
			resources:        []corev2.Resource{event, resource1},
			mockSelectorFunc: mockRegionSelector,
			want:             []corev2.Resource{event},
		},
		{
			name:             "event's check & entity labels",
			resources:        []corev2.Resource{event, resource1},
			mockSelectorFunc: labelSelector.Matches,
			want:             []corev2.Resource{event},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filter(tt.resources, tt.mockSelectorFunc); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
