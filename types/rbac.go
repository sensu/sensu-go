package types

const (
	// ResourceAll represents all possible resources
	ResourceAll = "*"
	// VerbAll represents all possible verbs
	VerbAll = "*"

	// GroupKind represents a group object in a subject
	GroupKind = "Group"
	// UserKind represents a user object in a subject
	UserKind = "User"
)

// ResourceMatches returns whether the specified requestedResource matches any
// of the rule resources
func (r Rule) ResourceMatches(requestedResource string) bool {
	for _, resource := range r.Resources {
		if resource == ResourceAll {
			return true
		}

		if resource == requestedResource {
			return true
		}
	}

	return false
}

// ResourceNameMatches returns whether the specified requestedResourceName
// matches any of the rule resources
func (r Rule) ResourceNameMatches(requestedResourceName string) bool {
	if len(requestedResourceName) == 0 {
		return true
	}

	for _, name := range r.ResourceNames {
		if name == requestedResourceName {
			return true
		}
	}

	return false
}

// VerbMatches returns whether the specified requestedVerb matches any of the
// rule verbs
func (r Rule) VerbMatches(requestedVerb string) bool {
	for _, verb := range r.Verbs {
		if verb == VerbAll {
			return true
		}

		if verb == requestedVerb {
			return true
		}
	}

	return false
}
