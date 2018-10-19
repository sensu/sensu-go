`freeze-api`
------------

### Introduction
The `freeze-api` tool is the official way to create public APIs for sensu-go.
Public APIs are created from internal APIs. Internal APIs are housed in
`github.com/sensu/sensu-go/internal/apis`. When the development team decides it
is time for an internal API to be publicized, then a public, versioned API will
be created in `github.com/sensu/sensu-go/apis`.


### Usage
```
$ freeze-api -h
Usage of freeze-api:
  -from string
    	Package to freeze
  -to string
    	Versioned package to create
```

#### Example Usage
```
freeze-api -from github.com/sensu/sensu-go/internal/apis/rbac -to github.com/sensu/sensu-go/apis/rbac/v1
```


### What does the `freeze-api` tool actually do?
The `freeze-api` tool has two key responsibilities.

1. Create the versioned package hierarchy based on the internal package hierarchy.
2. Generate conversion functions for the public API data types. This will allow
the data types to be converted to internal data types.


### How can developers create a public sensu-go API?
Follow these steps to create a public API for sensu-go.

1. Create one or more datatypes that embed `meta.TypeMeta` and `meta.ObjectMeta`
in the `internal/apis` package.
2. Run the `go-to-protobuf` tool. This needs to be formalized somewhat, but see
the `hack` directory for more information. This tool will generate a `.proto`
file and compile it to Go source code.
3. Implement any business logic required in the API.
4. Run the `freeze-api` tool on the internal package, creating a new versioned
package in `github.com/sensu/sensu-go/apis`. Use a kubernetes-style version
name for the package. If you aren't sure what this means, have a look at the
existing public APIs.
5. Run `go generate github.com/sensu/sensu-go/runtime/registry`. This will
register the newly created versioned types. (In fact, it will register all
of the types in the project that have embedded `meta.TypeMeta`.)


### What are converters?
Converters are functions that convert from a published API type to an internal
API type. They are patchable; that is, developers can redefine how types should
be converted. When making changes to an existing internal API, it may be
necessary to override one or more of the default converters, which only do
simple pointer conversion.

Here's an example of overriding a converter in the v1alpha1 rbac package:
```override.go
package v1alpha1

func init() {
  // Replace the existing conversion function variable with a new one.
  convert_Role_To_rbac_Role = override_convert_Role_To_rbac_Role
}

// A contrived example to demonstrate overloading a conversion function.
// In this example, a developer has identified the need to deprecate certain
// RBAC rule verbs.
func override_convert_Role_To_rbac_Role(from *Role, to *rbac.Role) {
  to.TypeMeta = from.TypeMeta
  to.ObjectMeta = from.ObjectMeta

  for _, r := range from.Rules {
    for _, verb := range r.Verbs {
      newVerbs := []string{}
      if verb != "deprecated" {
        newVerbs = append(newVerbs, verb)
      }
    }
  }
  newRule := r
  newRule.Verbs = newVerbs
  to.Rules = append(to.Rules, newRule)
}
```

In `func init`, the existing `convert_Role_To_rbac_Role` function variable
is replaced with a new one. The Role's Convert method will call the new
function at runtime instead of the old one. Developers should take care to
not perform overly expensive operations in conversions. Use reflection
sparingly, and don't do round-trip marshaling.


### Conclusion
The `freeze-api` tool, and the new API hierarchy, are being introduced in an
effort to standardize the package layout of sensu-go, while providing a way
to develop new APIs in a backwards-compatible manner. While developers should
expect some things to change in the future, the general approach to creating
APIs should be more or less consistent with what's described here.
