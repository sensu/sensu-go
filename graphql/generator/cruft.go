package generator

import "github.com/dave/jennifer/jen"

func newGroup() *jen.Group {
	// NOTE:
	// Group's separator field is not exported, so at this time this is the
	// only way to get a new group w/ the separator set to \n w/o forking.
	file := jen.NewFile("blah")
	return file.Group
}
