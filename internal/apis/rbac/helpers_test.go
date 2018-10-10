package rbac

import "testing"

func TestAPIGroupMatches(t *testing.T) {
	tests := []struct {
		name              string
		apiGroups         []string
		requestedAPIGroup string
		want              bool
	}{
		{
			name:              "empty rule api groups",
			requestedAPIGroup: "core.sensu.io",
			want:              false,
		},
		{
			name:              "all api groups",
			apiGroups:         []string{APIGroupAll},
			requestedAPIGroup: "core.sensu.io",
			want:              true,
		},
		{
			name:              "does not match",
			apiGroups:         []string{"core.sensu.io"},
			requestedAPIGroup: "rbac.authorization.sensu.io",
			want:              false,
		},
		{
			name:              "matches",
			apiGroups:         []string{"core.sensu.io"},
			requestedAPIGroup: "core.sensu.io",
			want:              true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := Rule{
				APIGroups: tc.apiGroups,
			}
			if got := r.APIGroupMatches(tc.requestedAPIGroup); got != tc.want {
				t.Errorf("Rule.APIGroupMatches() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRuleResourceMatches(t *testing.T) {
	tests := []struct {
		name              string
		resources         []string
		requestedResource string
		want              bool
	}{
		{
			name:              "empty rule resources",
			requestedResource: "checks",
			want:              false,
		},
		{
			name:              "all resources",
			resources:         []string{ResourceAll},
			requestedResource: "checks",
			want:              true,
		},
		{
			name:              "does not match",
			resources:         []string{"checks"},
			requestedResource: "events",
			want:              false,
		},
		{
			name:              "matches",
			resources:         []string{"checks", "events"},
			requestedResource: "events",
			want:              true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := Rule{
				Resources: tc.resources,
			}
			if got := r.ResourceMatches(tc.requestedResource); got != tc.want {
				t.Errorf("Rule.ResourceMatches() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRuleResourceNameMatches(t *testing.T) {
	tests := []struct {
		name                  string
		resourceNames         []string
		requestedResourceName string
		want                  bool
	}{
		{
			name: "empty rule resources",
			requestedResourceName: "checks",
			want: false,
		},
		{
			name:          "no name specified",
			resourceNames: []string{"foo"},
			want:          true,
		},
		{
			name:                  "does not match",
			resourceNames:         []string{"foo"},
			requestedResourceName: "bar",
			want: false,
		},
		{
			name:                  "matches",
			resourceNames:         []string{"foo", "bar"},
			requestedResourceName: "bar",
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := Rule{
				ResourceNames: tc.resourceNames,
			}
			if got := r.ResourceNameMatches(tc.requestedResourceName); got != tc.want {
				t.Errorf("Rule.ResourceNameMatches() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestRuleVerbMatches(t *testing.T) {
	tests := []struct {
		name          string
		verbs         []string
		requestedVerb string
		want          bool
	}{
		{
			name:          "empty rule verbs",
			requestedVerb: "get",
			want:          false,
		},
		{
			name:          "all verbs",
			verbs:         []string{VerbAll},
			requestedVerb: "get",
			want:          true,
		},
		{
			name:          "does not match",
			verbs:         []string{"create"},
			requestedVerb: "get",
			want:          false,
		},
		{
			name:          "matches",
			verbs:         []string{"create", "get"},
			requestedVerb: "get",
			want:          true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := Rule{
				Verbs: tc.verbs,
			}
			if got := r.VerbMatches(tc.requestedVerb); got != tc.want {
				t.Errorf("Rule.VerbMatches() = %v, want %v", got, tc.want)
			}
		})
	}
}
