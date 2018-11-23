package client

import (
	"net/url"
	"path"
)

func createBasePath(group, version, resType string) func(string, ...string) string {
	baseComponents := []string{"/api", group, version}

	return func(namespace string, pathComponents ...string) string {
		components := []string{}
		components = append(components, baseComponents...)

		// namespace
		if namespace != "" {
			components = append(components, "namespaces", namespace)
		}

		// resource type
		components = append(components, resType)

		// append given path components
		for _, path := range pathComponents {
			components = append(components, url.PathEscape(path))
		}

		return path.Join(components...)
	}
}
