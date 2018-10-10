package rbac

// APIGroupMatches returns whether the specified requestedAPIGroup matches any of
// the rule API groups
func (r Rule) APIGroupMatches(requestedAPIGroup string) bool {
	for _, group := range r.APIGroups {
		if group == APIGroupAll {
			return true
		}

		if group == requestedAPIGroup {
			return true
		}
	}

	return false
}

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
