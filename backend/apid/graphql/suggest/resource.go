package suggest

import corev2 "github.com/sensu/core/v2"

// Resource represents a Sensu resource
type Resource struct {
	Group      string
	Name       string
	Fields     []Field
	FilterFunc func(corev2.Resource) map[string]string
}

// LookupField uses given ref to find the appropriate field.
func (r *Resource) LookupField(ref RefComponents) Field {
	for _, f := range r.Fields {
		if f.Matches(ref.FieldPath) {
			return f
		}
	}
	return nil
}
