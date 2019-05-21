package filter

import (
	"errors"
	"strings"

	v2 "github.com/sensu/sensu-go/api/core/v2"
)

var (
	// KeylessStatementErr statement is missing a key
	KeylessStatementErr = errors.New("filters must have the format KEY:VAL")

	// FilterNotFoundErr could not match a filter for the given key
	FilterNotFoundErr = errors.New("no filter could be matched with the given statement")
)

// Match a given resource
type Matcher func(v2.Resource) bool

// Filter configures a new Matcher given a statement and a fields func.
type Filter func(string, FieldsFunc) (Matcher, error)

// FieldsFunc represents the function to retrieve fields about a given resource
type FieldsFunc func(resource v2.Resource) map[string]string

const (
	// separator character used to separate the key and value
	statementSeparator = ":"
)

// Compile matcher from given statements, filters and fields.
func Compile(statements []string, filters map[string]Filter, fieldsFn FieldsFunc) (Matcher, error) {
	matchers := []Matcher{}
	for _, s := range statements {
		ss := strings.SplitN(s, statementSeparator, 2)
		if len(ss) != 2 {
			return nil, KeylessStatementErr
		}
		k, v := ss[0], ss[1]
		f, ok := filters[k]
		if !ok {
			return nil, FilterNotFoundErr
		}
		matcher, err := f(v, fieldsFn)
		if err != nil {
			return nil, err
		}
		matchers = append(matchers, matcher)
	}

	return func(res v2.Resource) bool {
		for _, matches := range matchers {
			if !matches(res) {
				return false
			}
		}
		return true
	}, nil
}
