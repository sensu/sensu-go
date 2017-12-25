package generator

import (
	"errors"
	"io/ioutil"
	"path/filepath"

	"github.com/jamesdphillips/graphql/language/ast"
	"github.com/jamesdphillips/graphql/language/parser"
)

// GraphQLFile encapsulates parsed document and filepath.
type GraphQLFile struct {
	path string
	ast  *ast.Document
}

// ParseFile parses given path to GraphQL file returning parsed AST and error if
// any occurred while parsing.
func ParseFile(path string) (*GraphQLFile, error) {
	bin, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	params := parser.ParseParams{Source: string(bin)}
	ast, err := parser.Parse(params)
	if err != nil {
		return nil, err
	}

	return &GraphQLFile{
		path: path,
		ast:  ast,
	}, nil
}

// Filename returns the file's name
func (file *GraphQLFile) Filename() string {
	return filepath.Base(file.path)
}

// Definitions returns all definition nodes found in document root.
func (file *GraphQLFile) Definitions() []ast.Node {
	return file.ast.Definitions
}

// Validate returns an error if given file does not appear to be GraphQL Schema
// Definition Language (SDL).
func (file *GraphQLFile) Validate() error {
	defs := file.Definitions()
	for _, def := range defs { // TODO: Check more than top level?
		switch def.(type) {
		case *ast.OperationDefinition:
			return errors.New("file should not define any operations")
		case *ast.FragmentDefinition:
			return errors.New("file should not define any fragments")
		default:
			// TODO: Validate that top level are unique.
			// TODO: Validate that unsupported directives are not given.
			// TODO: Validate names; eg. types should use upper CamelCase where as
			//       fields should use lower camelCase.
			continue
		}
	}
	return nil
}
