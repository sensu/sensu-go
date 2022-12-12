package filter

import (
	v2 "github.com/sensu/core/v2"
)

type StringMatcher func(v2.Resource, string) bool

func String(fn StringMatcher) Filter {
	return func(val string, _ FieldsFunc) (Matcher, error) {
		return func(res v2.Resource) bool {
			return fn(res, val)
		}, nil
	}
}
