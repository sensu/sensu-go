package globalid

import (
	"errors"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

//
// IDs
//

// Components describes the components of a global identifier.
//
// When represented as a string the ID appears in the follwoing format, parens
// denote optional components.
//
//   srn:resource:(?ns:)(?resourceType/)uniqueComponents?extras=values&more=kvs
//
// Example global IDs
//
//   srn:entities:default:selene.local
//   srn:events:sales:check/aG93ZHkgYnVkCg==
//   srn:checks:auto:disk-full
//   srn:users:deanlearner
//
type Components interface {
	// Resource definition associated with this ID.
	Resource() string

	// SetResource value.
	SetResource(string)

	// Namespace is the name of the namespace the resource belongs to.
	Namespace() string

	// ResourceType is a optional element that describes any sort of sub-type of
	// the resource.
	ResourceType() string

	// UniqueComponent is a string that uniquely identify a resource; often times
	// this is the resource's name.
	UniqueComponent() string

	// Extras returns interface with access to extra values
	Extras() Values

	// String return string representation of ID
	String() string
}

type Values interface {
	// Get gets the first value associated with the given key. If there are no
	// values associated with the key, Get returns the empty string. To access
	// multiple values, use the map directly.
	Get(key string) string

	// Set sets the key to value. It replaces any existing values.
	Set(key, val string)

	// Clear removes all keys
	Clear()
}

// StandardComponents describes the standard components of a global identifier.
type StandardComponents struct {
	resource        string
	namespace       string
	resourceType    string
	uniqueComponent string
	extras          url.Values
}

// String returns the string representation of the global ID.
func (id *StandardComponents) String() string {
	uniqueComponent := url.PathEscape(id.uniqueComponent)
	nameComponents := append([]string{id.resourceType}, uniqueComponent)
	nameComponents = omitEmpty(nameComponents)
	pathComponents := omitEmpty([]string{
		id.resource,
		id.namespace,
	})

	// srn:{pathComponents}:{nameComponents}
	str := "srn:" + strings.Join(pathComponents, ":") +
		":" + strings.Join(nameComponents, "/")

	if len(id.extras) > 0 {
		str += "?" + id.extras.Encode()
	}
	return str
}

// Resource definition associated with this ID.
func (id *StandardComponents) Resource() string {
	return id.resource
}

// Resource definition associated with this ID.
func (id *StandardComponents) SetResource(str string) {
	id.resource = str
}

// Namespace is the name of the namespace the resource belongs to.
func (id *StandardComponents) Namespace() string {
	return id.namespace
}

// ResourceType is a optional element that describes any sort of sub-type of
// the resource.
func (id *StandardComponents) ResourceType() string {
	return id.resourceType
}

// UniqueComponent is a string that uniquely identify a resource; often times
// this is the resource's name.
func (id *StandardComponents) UniqueComponent() string {
	return id.uniqueComponent
}

// Extra returns the extra value associated with the given key.
func (id *StandardComponents) Extras() Values {
	if id.extras == nil {
		id.extras = url.Values{}
	}
	return &dict{id.extras}
}

// Parse takes a global ID string, decodes it and returns it's components.
func Parse(gid string) (*StandardComponents, error) {
	id := &StandardComponents{}

	// Clip extras from end of globalid, eg. srn:resource:name?entity=proxy
	//                                                        ^^^^^^^^^^^^^
	if i := strings.LastIndexByte(gid, '?'); i != -1 {
		v, err := url.ParseQuery(gid[i+1:])
		if err != nil {
			logrus.WithField("component", "graphql/globalid").WithError(err).Warn("unable to parse query params")
		}
		id.extras = v
		gid = gid[:i]
	}

	pathComponents := strings.SplitN(gid, ":", 4)

	// Should be at least srn:resource:name
	if len(pathComponents) < 3 {
		return id, errors.New("given global ID does not appear valid")
	}

	if pathComponents[0] != "srn" {
		return id, errors.New("given string does not appear to be a Sensu global ID")
	}

	// Pop the resource from the path components, eg. srn:resource:ns:type/name
	//                                                    ^^^^^^^^
	id.resource = pathComponents[1]
	pathComponents = pathComponents[2:]

	// Pop the name components from the path components, eg. ns:type/name
	//                                                          ^^^^^^^^^
	nameComponents := strings.Split(pathComponents[len(pathComponents)-1], "/")
	pathComponents = pathComponents[0 : len(pathComponents)-1]

	// If present pop the ns from the path components, eg. ns
	//                                                     ^^
	if len(pathComponents) > 0 {
		id.namespace = pathComponents[0]
	}

	// If present pop the type from the name components, eg. type/my-great-check
	//                                                       ^^^^
	if len(nameComponents) > 1 {
		id.resourceType = nameComponents[0]
		nameComponents = nameComponents[1:]
	}

	// Pop the remaining element from the name components, eg. my-great-check
	//                                                         ^^^^^^^^^^^^^^
	id.uniqueComponent, _ = url.PathUnescape(nameComponents[0])

	return id, nil
}

func omitEmpty(in []string) (out []string) {
	for _, n := range in {
		if n != "" {
			out = append(out, n)
		}
	}

	return
}

type dict struct {
	url.Values
}

func (d *dict) Clear() {
	for key := range d.Values {
		d.Values.Del(key)
	}
}
