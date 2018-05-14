package main

import (
	"path"
	"strings"
)

type (
	Type struct {
		packageName string
		name        string
	}

	Scope struct {
		converter   *Converter
		packageName string
		path        []string
		types       map[string]*Type
		imports     map[string]*Type
		children    map[string]*Scope
	}
)

func NewScope(converter *Converter) *Scope {
	return &Scope{
		converter: converter,
		path:      make([]string, 0, 5),
		types:     make(map[string]*Type),
		imports:   make(map[string]*Type),
		children:  make(map[string]*Scope),
	}
}

func (s *Scope) Fork(name string) *Scope {
	p := make([]string, len(s.path))

	copy(p, s.path)

	childScope := &Scope{
		converter:   s.converter,
		packageName: s.packageName,
		types:       s.types, // share types collection
		path:        append(p, name),
		children:    make(map[string]*Scope),
	}

	s.children[name] = childScope

	return childScope
}

func (s *Scope) SetPackageName(name string) {
	s.packageName = s.converter.PackageName(strings.Split(name, "."))
}

func (s *Scope) AddLocalType(name string) {
	typeName := s.converter.OriginalTypeName(s, name)

	_, ok := s.types[typeName]

	if ok == false {
		s.types[typeName] = &Type{
			packageName: s.packageName,
			name:        s.converter.NewTypeName(s, name),
		}
	}
}

func (s *Scope) AddImportedType(filename string) {
	dir := path.Dir(filename)
	name := strings.Replace(path.Base(filename), ".proto", "", -1)
	ref := strings.Replace(dir, "/", ".", -1)

	var packageName string

	if dir == "." {
		packageName = s.packageName
	} else {
		packageName = s.converter.PackageName(strings.Split(dir, "/"))
	}

	s.imports[ref] = &Type{packageName, name}
}

func (s *Scope) ResolveTypeName(ref string) string {
	builtin, ok := BUILTINS[ref]

	if ok == true {
		return builtin
	}

	// try to find one in a global scope
	scoped, ok := s.types[ref]

	if ok == true {
		return scoped.name
	}

	// try to find one among nested types
	nested, ok := s.types[s.converter.OriginalTypeName(s, ref)]

	if ok == true {
		return nested.name
	}

	var foundInChildren string

	for _, childScope := range s.children {
		res := childScope.ResolveTypeName(ref)

		if res != ref {
			foundInChildren = res
			break
		}
	}

	if foundInChildren != "" {
		return foundInChildren
	}

	// if we are still here, probably it's an imported type

	// if type does not contain "." it means it's from the same package
	if strings.Contains(ref, ".") == false {
		// from the same package
		imported, ok := s.imports["."]

		if ok == true {
			return imported.packageName + ref
		}
	} else {
		// if it has "." it means it's from other package
		parts := strings.Split(ref, ".")

		var tail string

		for idx, segment := range parts {
			if tail == "" {
				tail = segment
			} else {
				tail += "." + segment
			}

			imported, ok := s.imports[tail]

			if ok == true {
				return imported.packageName + strings.Join(parts[idx+1:], "")
			}
		}
	}

	return ref
}
