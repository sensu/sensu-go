package selector

import (
	"strings"
)

// Selector represents a field or label selector that declares one or more
// operations.
type Selector struct {
	Operations []Operation
}

// Matches returns the logical intersection of the evaluations of each of the
// operations in s.
func (s *Selector) Matches(set map[string]string) bool {
	for i := range s.Operations {
		if matches := matches(s.Operations[i], set); !matches {
			return false
		}
	}

	return true
}

// match determines if an operation matches the given set
func matches(r Operation, set map[string]string) bool {
	switch r.Operator {
	case InOperator:
		// Verify if we have an l-value in the r-values.
		// e.g. linux in (check.subscriptions)
		if hasKeysInValues(set, r.RValues) {
			return hasValue(r.LValue, split(set[r.RValues[0]]))
		}
		// We are not dealing with a key in the values so we can follow the same
		// logic as the equal operator
		fallthrough
	case DoubleEqualSignOperator:
		// Make sure the r-value set has the specified l-value
		if !hasKey(set, r.LValue) {
			return false
		}
		// Make sure the r-value set's value for the operation's l-value exists in the
		// operation r-values.
		return hasValue(set[r.LValue], r.RValues)
	case NotInOperator:
		// Verify if we have an l-value in the r-values.
		// e.g. linux notin (check.subscriptions)
		if hasKeysInValues(set, r.RValues) {
			return !hasValue(r.LValue, split(set[r.RValues[0]]))
		}
		fallthrough
	case NotEqualOperator:
		// Make sure the r-value set has the specified l-value.
		if !hasKey(set, r.LValue) {
			return true
		}
		//  Make sure the set's value for the operation's l-value does not exists in
		//  the operation r-values.
		return !hasValue(set[r.LValue], r.RValues)
	case MatchesOperator:
		// Make sure the r-value set has the specified l-value
		if !hasKey(set, r.LValue) {
			return false
		}
		//  Make sure the set's value for the operation's l-value matches
		//  the operation r-values
		return matchesValue(set[r.LValue], r.RValues)
	default:
		return false
	}
}

// hasKey determines if the given set has a key with the specified name
func hasKey(set map[string]string, key string) bool {
	_, ok := set[key]
	return ok
}

// hasValue determines if the set value exists in the operation values
func hasValue(value string, values []string) bool {
	for i := range values {
		if values[i] == value {
			return true
		}
	}

	return false
}

// matchesValue determines if the set value matches in the operation values
func matchesValue(value string, values []string) bool {
	for i := range values {
		if strings.Contains(value, values[i]) {
			return true
		}
	}

	return false
}

// hasKeysInValues determines if the values contains an actual key of the set
func hasKeysInValues(set map[string]string, values []string) bool {
	// We only support a single value in the array
	// e.g. linux in (check.subscriptions)
	if len(values) != 1 {
		return false
	}
	if !hasKey(set, values[0]) {
		return false
	}
	return true
}

// split slices input into substrings and remove any whitespaces
func split(input string) []string {
	input = strings.Trim(input, "[]")
	s := strings.Split(input, ",")
	for i := range s {
		s[i] = strings.TrimSpace(s[i])
	}
	return s
}

// Merge merges many selectors into one mega selector!
func Merge(selectors ...*Selector) *Selector {
	var selector Selector
	for _, s := range selectors {
		if s == nil {
			continue
		}
		selector.Operations = append(selector.Operations, s.Operations...)
	}
	return &selector
}
