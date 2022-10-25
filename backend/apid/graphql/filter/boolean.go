package filter

import (
	"strconv"
	"strings"

	v2 "github.com/sensu/core/v2"
)

type BooleanMatcher func(v2.Resource, bool) bool

func Boolean(fn BooleanMatcher) Filter {
	return func(val string, _ FieldsFunc) (Matcher, error) {
		v, err := strconv.ParseBool(strings.TrimSpace(val))
		if err != nil {
			return nil, err
		}

		return func(res v2.Resource) bool {
			return fn(res, v)
		}, nil
	}
}
