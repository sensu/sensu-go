package filter

import (
	"strconv"
	"strings"

	corev3 "github.com/sensu/core/v3"
)

type BooleanMatcher func(corev3.Resource, bool) bool

func Boolean(fn BooleanMatcher) Filter {
	return func(val string, _ FieldsFunc) (Matcher, error) {
		v, err := strconv.ParseBool(strings.TrimSpace(val))
		if err != nil {
			return nil, err
		}

		return func(res corev3.Resource) bool {
			return fn(res, v)
		}, nil
	}
}
