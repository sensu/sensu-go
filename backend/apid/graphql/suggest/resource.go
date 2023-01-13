package suggest

import corev3 "github.com/sensu/core/v3"

// Resource represents a Sensu resource
type Resource struct {
	Group      string
	Name       string
	Fields     []Field
	FilterFunc func(corev3.Resource) map[string]string
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
