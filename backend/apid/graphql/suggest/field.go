package suggest

import (
	corev3 "github.com/sensu/core/v3"
)

var (
	NameField = &CustomField{
		Name: "name",
		FieldFunc: func(res corev3.Resource) []string {
			return []string{res.GetMetadata().Name}
		},
	}
	LabelsField = &MapField{
		Name: "labels",
		FieldFunc: func(res corev3.Resource) map[string]string {
			return res.GetMetadata().Labels
		},
	}
)

type Field interface {
	Matches(string) bool
	Value(corev3.Resource, string) []string
}

type CustomField struct {
	Name      string
	FieldFunc func(corev3.Resource) []string
}

func (f *CustomField) Matches(path string) bool {
	return f.Name == path
}

func (f *CustomField) Value(res corev3.Resource, path string) []string {
	return f.FieldFunc(res)
}

type MapField struct {
	Name      string
	FieldFunc func(corev3.Resource) map[string]string
}

func (f *MapField) Matches(path string) bool {
	return startsWith(path, f.Name)
}

func (f *MapField) Value(res corev3.Resource, path string) []string {
	key := trimSeg(path, f.Name+"/")
	fld := f.FieldFunc(res)
	if key == "" {
		return collectKeys(fld)
	}
	val, ok := fld[key]
	if ok {
		return []string{val}
	}
	return []string{}
}

type ObjectField struct {
	Name   string
	Fields []Field
}

func (f *ObjectField) Matches(path string) bool {
	if !startsWith(path, f.Name) {
		return false
	}
	path = trimSeg(path, f.Name+"/")
	for _, n := range f.Fields {
		if n.Matches(path) {
			return true
		}
	}
	return false
}

func (f *ObjectField) Value(res corev3.Resource, path string) []string {
	path = trimSeg(path, f.Name+"/")
	for _, n := range f.Fields {
		if n.Matches(path) {
			return n.Value(res, path)
		}
	}
	return []string{}
}

func startsWith(path, seg string) bool {
	if len(path) < len(seg) {
		return false
	} else if len(path) == len(seg) {
		return path == seg
	}
	seg += "/"
	return path[:len(seg)] == seg
}

func trimSeg(path, seg string) string {
	if len(path) < len(seg) {
		return ""
	}
	return path[len(seg):]
}

func collectKeys(labels map[string]string) (vals []string) {
	for key := range labels {
		vals = append(vals, key)
	}
	return
}
