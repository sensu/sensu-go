package graphql

import (
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
	"github.com/sensu/sensu-go/backend/selector"
)

// GlobalFilters are filters that are applied to all resolvers that accept
// filter statements.
var GlobalFilters = DefaultGlobalFilters()

// DefaultGlobalFilters returns the default set of global filters.
func DefaultGlobalFilters() map[string]filter.Filter {
	return map[string]filter.Filter{
		"labelSelector": labelSelector,
		"fieldSelector": fieldSelector,
	}
}

func labelSelector(v string, _ filter.FieldsFunc) (filter.Matcher, error) {
	selector, err := selector.ParseLabelSelector(v)
	if err != nil {
		// TODO:
		//
		// It would be much more ideal to surface any parse errors to the user, however,
		// currently the web client cannot handle the errors returned by the GraphQL service
		// well. As a work around we are currently logging the error and filterting out all
		// resources.
		//
		logger.Error(err)
		return matchNone, nil
	}

	return func(res corev3.Resource) bool {
		if selector.Matches(res.GetMetadata().Labels) {
			return true
		}

		// If we have an event, we also want to inspect the check and entitiy
		// labels. Events are never wrapped in the store so we do not need this
		// condition in the switch's case below.
		if event, ok := res.(*corev2.Event); ok {
			// We don't need to merge the labels here because the selector only
			// contains a single requirement
			if selector.Matches(event.Check.GetLabels()) || selector.Matches(event.Entity.GetLabels()) {
				return true
			}
		}

		return false
	}, nil
}

func fieldSelector(v string, fields filter.FieldsFunc) (filter.Matcher, error) {
	selector, err := selector.ParseFieldSelector(v)
	if err != nil {
		// TODO:
		//
		// It would be much more ideal to surface any parse errors to the user, however,
		// currently the web client cannot handle the errors returned by the GraphQL service
		// well. As a work around we are currently logging the error and filterting out all
		// resources.
		//
		logger.Error(err)
		return matchNone, nil
	}

	return func(res corev3.Resource) bool {
		return selector.Matches(fields(res))
	}, nil
}

func matchNone(_ corev3.Resource) bool {
	return false
}
