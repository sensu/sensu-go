package graphql

import (
	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/filter"
)

// EventFilters returns collection of filters used for matching resources.
func EventFilters() map[string]filter.Filter {
	filters := map[string]filter.Filter{
		// status:passing | status:warning | status:unknown | status:incident
		"status": filter.String(func(res v2.Resource, v string) bool {
			status := res.(*v2.Event).Check.Status
			switch v {
			case "passing":
				return status == 0
			case "warning":
				return status == 1
			case "critical":
				return status == 2
			case "unknown":
				return status > 2
			case "incident":
				return status > 0
			default:
				return false
			}
		}),
		// check:check-disk
		"check": filter.String(func(res v2.Resource, v string) bool {
			return res.(*v2.Event).Check.Name == v
		}),
		// entity:server
		"entity": filter.String(func(res v2.Resource, v string) bool {
			return res.(*v2.Event).Entity.Name == v
		}),
		// silenced:true | silenced:false
		"silenced": filter.Boolean(func(res v2.Resource, v bool) bool {
			return (len(res.(*v2.Event).Check.Silenced) > 0) == v
		}),
	}

	// merge global filters
	for k, f := range GlobalFilters {
		filters[k] = f
	}

	return filters
}
