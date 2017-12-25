package util

import "github.com/graphql-go/graphql"

//
// == Forward ==
//
// A pretty frustrating aspect of the type definitions is that they can be
// somewhat order dependent. For instance, if my Dog references Breeds and
// implements the Pet interface then I need to make sure those are loaded
// first.
//
// There are tricks to get around this (using a reference to the type and
// thunks) but the are not 100% perfect either and create a lot of needless
// code. This becomes even more tricky when generating the types.
//
// As such to get around this we generate a mock type when generating the
// object configuration; this mock only refers to the unique name of the
// component it is referencing and is replace at the time the GraphQL service
// is invoked.

type mockType struct{ name string }

func (o *mockType) Name() string        { return o.name }
func (o *mockType) Description() string { return "" }
func (o *mockType) String() string      { return o.name }
func (o *mockType) Error() error        { return nil }

// OutputType mocks a type (Object, Scalar, Enum, etc.)
func OutputType(name string) graphql.Output {
	return &mockType{name}
}

// InputType mocks a type (InputObject, Scalar, Enum, etc.)
func InputType(name string) graphql.Input {
	return &mockType{name}
}

// Interface mocks an interface
func Interface(name string) *graphql.Interface {
	// Unlike fields which simply require that something that implements the
	// Output interface is present object types require that references to
	// interfaces are given to config.
	//
	// Feels a bit brittle but simplest solution at this time.
	return &graphql.Interface{PrivateName: name}
}

// Object mocks an interface
func Object(name string) *graphql.Object {
	// Unlike fields which simply require that something that implements the
	// Output interface is present, schema & union types require that references
	// to object are given to config.
	//
	// Feels a bit brittle but simplest solution at this time.
	return &graphql.Object{PrivateName: name}
}
