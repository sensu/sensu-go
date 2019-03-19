package filter

import (
	"github.com/sensu/sensu-go/api/core/v2"
	"k8s.io/apimachinery/pkg/fields"
)

// labeSelector is PoC for filtering based on a label
func labelSelector() {

}

// eventsField returns a field set that represents the available fields
func eventsField(event *v2.Event) fields.Set {
	eventsFieldSet := make(fields.Set, 3)
	eventsFieldSet["event.check.name"] = event.Check.ObjectMeta.Name
	eventsFieldSet["event.entity.name"] = event.Entity.ObjectMeta.Name
	eventsFieldSet["event.check.status"] = string(event.Check.Status)
	return eventsFieldSet
}
