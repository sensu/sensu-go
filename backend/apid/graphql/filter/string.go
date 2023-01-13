package filter

import corev3 "github.com/sensu/core/v3"

type StringMatcher func(corev3.Resource, string) bool

func String(fn StringMatcher) Filter {
	return func(val string, _ FieldsFunc) (Matcher, error) {
		return func(res corev3.Resource) bool {
			return fn(res, val)
		}, nil
	}
}
