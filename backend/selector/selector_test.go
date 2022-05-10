package selector

import (
	"testing"
)

func TestSelector_Matches(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		set     map[string]string
		want    bool
		wantErr bool
	}{
		{
			name:  "equal matches",
			input: "object.name == foo",
			set:   map[string]string{"object.name": "foo"},
			want:  true,
		},
		{
			name:  "equal doesn't match",
			input: "object.name == bar",
			set:   map[string]string{"object.name": "foo"},
			want:  false,
		},
		{
			name:  "in matches",
			input: "object.name in [foo,bar]",
			set:   map[string]string{"object.name": "foo"},
			want:  true,
		},
		{
			name:  "in doesn't match",
			input: "object.name in [bar,baz]",
			set:   map[string]string{"object.name": "foo"},
			want:  false,
		},
		{
			name:  "key doesn't exist in equal",
			input: "object.name == foo",
			set:   map[string]string{"object.namespace": "acme"},
			want:  false,
		},
		{
			name:  "notequal matches",
			input: "object.name != bar",
			set:   map[string]string{"object.name": "foo"},
			want:  true,
		},
		{
			name:  "notequal doesn't match",
			input: "object.name != foo",
			set:   map[string]string{"object.name": "foo"},
			want:  false,
		},
		{
			name:  "notin matches",
			input: "object.name notin [bar,baz]",
			set:   map[string]string{"object.name": "foo"},
			want:  true,
		},
		{
			name:  "notin doesn't match",
			input: "object.name notin [foo,bar]",
			set:   map[string]string{"object.name": "foo"},
			want:  false,
		},
		{
			name:  "key doesn't exist in notequal",
			input: "object.name != foo",
			set:   map[string]string{"object.namespace": "acme"},
			want:  true,
		},
		{
			name:  "multiple requirements match",
			input: "object.name != foo && object.namespace == acme",
			set:   map[string]string{"object.name": "bar", "object.namespace": "acme"},
			want:  true,
		},
		{
			name:  "multiple requirements do not match",
			input: "object.name != foo && object.namespace == acme",
			set:   map[string]string{"object.name": "bar", "object.namespace": "dev"},
			want:  false,
		},
		{
			name:  "object key within array with in operator match",
			input: "bar in object.subscriptions",
			set:   map[string]string{"object.subscriptions": "[foo, bar]"},
			want:  true,
		},
		{
			name:  "object key within array with in operator does not match",
			input: "qux in object.subscriptions",
			set:   map[string]string{"object.subscriptions": "[foo,bar]"},
			want:  false,
		},
		{
			name:  "object key within array with notin operator match",
			input: "qux notin object.subscriptions",
			set:   map[string]string{"object.subscriptions": "[foo,bar]"},
			want:  true,
		},
		{
			name:  "object key within array with notin operator does not match",
			input: "foo notin object.subscriptions",
			set:   map[string]string{"object.subscriptions": "[foo,bar]"},
			want:  false,
		},
		{
			name:  "object key within array does not exist",
			input: "bar in object.namespace",
			set:   map[string]string{"object.subscriptions": "[foo, bar]"},
			want:  false,
		},
		{
			name:  "string value",
			input: "region == \"us-west-1\"",
			set:   map[string]string{"region": "us-west-1"},
			want:  true,
		},
		{
			name:  "string value with single quotes",
			input: "region == 'us-west-1'",
			set:   map[string]string{"region": "us-west-1"},
			want:  true,
		},
		{
			name:  "boolean value",
			input: "object.publish == true",
			set:   map[string]string{"object.publish": "true"},
			want:  true,
		},
		{
			name:  "matches doesn't match",
			input: "object.name matches bar",
			set:   map[string]string{"object.name": "foo"},
			want:  false,
		},
		{
			name:  "matches returns a match",
			input: "object.name matches fo",
			set:   map[string]string{"object.name": "foo"},
			want:  true,
		},
		{
			name:  "nil map with equality",
			input: "foo == 'bar'",
			set:   nil,
			want:  false,
		},
		{
			name:  "nil map with inequality",
			input: "object.foo != 'bar'",
			set:   nil,
			want:  true,
		},
		{
			name:  "nil map with matches",
			input: "foo matches bar",
			set:   nil,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Selector.Matches error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got := s.Matches(tt.set); got != tt.want {
				t.Errorf("Selector.Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}
