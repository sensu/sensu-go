package main

import "strings"

type Converter struct {
	noPrefix bool
	pkgAliases map[string]string
}

func (c *Converter) NewTypeName(scope *Scope, name string) string {
	return scope.packageName + strings.Join(scope.path, "") + name
}

func (c *Converter) OriginalTypeName(scope *Scope, name string) string {
	switch len(scope.path) {
	case 0:
		return name
	case 1:
		return scope.path[0] + "." + name
	default:
		return strings.Join(scope.path, ".") + "." + name
	}
}

func (c *Converter) PackageName(parts []string) string {
	var name string

	if c.noPrefix == true {
		return name
	}

	if len(c.pkgAliases) > 0 {
		alias, exists := c.pkgAliases[strings.Join(parts, ".")]

		if exists == true {
			return alias
		}
	}

	for _, segment := range parts {
		name += strings.Title(segment)
	}

	return name
}
