package client

import (
	"net/url"
	"path"
)

const coreAPIGroup = "core"
const coreAPIVersion = "v2"

// CreateBasePath ...
func CreateBasePath(group, version string, resType ...string) func(...string) string {
	fn := createNSBasePath(group, version, resType...)

	return func(components ...string) string {
		// first arg "" represents that no namespace has been specified
		return fn("", components...)
	}
}

// Returns an anonymous function that when given a namespace and additional path
// components returns a full path to a endpoint. If namespace is an empty string
// namespaces prefix is omitted from the resulting path.
//
//    pathFn := createNSBasePath("core", "v1", "users")
//    pathFn("sensu") // /api/core/v1/namespaces/sensu/users
//    pathFn("sensu", "frank-west") // /api/core/v1/namespaces/sensu/users/frank-west
//    pathFn("") // /api/core/v1/users
//
//    pathFn = createNSBasePath("core", "v1", "rbac", "rules")
//    pathFn("sensu", "admin") // /api/core/v1/namespaces/sensu/rbac/rules/admin
//    pathFn("") // /api/core/v1/namespaces/sensu/rbac/rules
//
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
