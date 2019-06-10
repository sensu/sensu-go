package suggest

import (
	"path"
	"regexp"
)

var (
	nsRe = regexp.MustCompile("{namespace}")
)

// Resource represents a Sensu resource
type Resource struct {
	Group  string
	Name   string
	Path   string
	Fields []Field
}

// URIPath given a namespace returns the API path used to get/list/put/delete
// the resource.
func (r *Resource) URIPath(ns string) string {
	if r.Path != "" {
		return nsRe.ReplaceAllString(r.Path, ns)
	}
	if r.Group == "core/v2" {
		return path.Join("/", "api", r.Group, "namespaces", ns, r.Name)
	}
	return path.Join("/", "api", r.Group, r.Name)
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
