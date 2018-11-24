package client

import (
	"net/url"
	"path"
)

const coreAPIGroup = "core"
const coreAPIVersion = "v2"

func createBasePath(group, version string, resType ...string) func(...string) string {
	fn := createNSBasePath(group, version, resType...)

	return func(components ...string) string {
		// first arg "" represents that no namespace has been specified
		return fn("", components...)
	}
}

func createNSBasePath(group, version string, resType ...string) func(string, ...string) string {
	baseComponents := []string{"/api", group, version}

	return func(namespace string, pathComponents ...string) string {
		components := []string{}
		components = append(components, baseComponents...)

		// namespace
		if namespace != "" {
			components = append(components, "namespaces", namespace)
		}

		// resource type
		components = append(components, resType...)

		// append given path components
		for _, path := range pathComponents {
			components = append(components, url.PathEscape(path))
		}

		return path.Join(components...)
	}
}
