package storage

import "testing"

type keyTest struct {
	Name        string
	Keyable     Keyable
	KeyerMethod func() string
	Expected    string
}

type keyableFixture struct {
	kind      string
	group     string
	version   string
	namespace string
	name      string
}

func (f keyableFixture) GetKind() string {
	return f.kind
}

func (f keyableFixture) GetGroup() string {
	return f.group
}

func (f keyableFixture) GetVersion() string {
	return f.version
}

func (f keyableFixture) GetName() string {
	return f.name
}

func (f keyableFixture) GetNamespace() string {
	return f.namespace
}

type noNamespaceKeyableFixture struct {
	keyableFixture
}

func (noNamespaceKeyableFixture) NoNamespace() {}

func TestRESTKey(t *testing.T) {
	fixture := keyableFixture{
		kind:      "frogs",
		group:     "muppets",
		version:   "v1",
		namespace: "default",
		name:      "kermit",
	}
	noNamespaceFixture := noNamespaceKeyableFixture{fixture}
	tests := []keyTest{
		{
			Name:        "RESTKey GET",
			KeyerMethod: RESTKey{fixture}.GetPath,
			Expected:    "/apis/muppets/v1/namespaces/default/frogs/kermit",
		},
		{
			Name:        "RESTKey PUT",
			KeyerMethod: RESTKey{fixture}.PutPath,
			Expected:    "/apis/muppets/v1/namespaces/default/frogs/kermit",
		},
		{
			Name:        "RESTKey PATCH",
			KeyerMethod: RESTKey{fixture}.PatchPath,
			Expected:    "/apis/muppets/v1/namespaces/default/frogs/kermit",
		},
		{
			Name:        "RESTKey DELETE",
			KeyerMethod: RESTKey{fixture}.DeletePath,
			Expected:    "/apis/muppets/v1/namespaces/default/frogs/kermit",
		},
		{
			Name:        "RESTKey List",
			KeyerMethod: RESTKey{fixture}.ListPath,
			Expected:    "/apis/muppets/v1/namespaces/default/frogs",
		},
		{
			Name:        "RESTKey POST",
			KeyerMethod: RESTKey{fixture}.PostPath,
			Expected:    "/apis/muppets/v1/namespaces/default/frogs",
		},
		{
			Name:        "RESTKey GET (no namespace)",
			KeyerMethod: RESTKey{noNamespaceFixture}.GetPath,
			Expected:    "/apis/muppets/v1/frogs/kermit",
		},
		{
			Name:        "RESTKey PUT (no namespace)",
			KeyerMethod: RESTKey{noNamespaceFixture}.PutPath,
			Expected:    "/apis/muppets/v1/frogs/kermit",
		},
		{
			Name:        "RESTKey PATCH (no namespace)",
			KeyerMethod: RESTKey{noNamespaceFixture}.PatchPath,
			Expected:    "/apis/muppets/v1/frogs/kermit",
		},
		{
			Name:        "RESTKey DELETE (no namespace)",
			KeyerMethod: RESTKey{noNamespaceFixture}.DeletePath,
			Expected:    "/apis/muppets/v1/frogs/kermit",
		},
		{
			Name:        "RESTKey List (no namespace)",
			KeyerMethod: RESTKey{noNamespaceFixture}.ListPath,
			Expected:    "/apis/muppets/v1/frogs",
		},
		{
			Name:        "RESTKey POST (no namespace)",
			KeyerMethod: RESTKey{noNamespaceFixture}.PostPath,
			Expected:    "/apis/muppets/v1/frogs",
		},
		{
			Name:        "StorageKey Path",
			KeyerMethod: StorageKey{fixture}.Path,
			Expected:    "/sensu.io/muppets/v1/default/frogs/kermit",
		},
		{
			Name:        "StorageKey PrefixPath",
			KeyerMethod: StorageKey{fixture}.PrefixPath,
			Expected:    "/sensu.io/muppets/v1/default/frogs",
		},
		{
			Name:        "StorageKey Path (no namespace)",
			KeyerMethod: StorageKey{noNamespaceFixture}.Path,
			Expected:    "/sensu.io/muppets/v1/frogs/kermit",
		},
		{
			Name:        "StorageKey PrefixPath (no namespace)",
			KeyerMethod: StorageKey{noNamespaceFixture}.PrefixPath,
			Expected:    "/sensu.io/muppets/v1/frogs",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			if got, want := test.KeyerMethod(), test.Expected; got != want {
				t.Fatalf("bad result: got %q, want %q", got, want)
			}
		})
	}
}
