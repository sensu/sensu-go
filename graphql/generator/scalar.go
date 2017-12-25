package generator

import (
	"fmt"
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genScalar(node *ast.ScalarDefinition) jen.Code {
	code := newGroup()
	name := node.GetName().Value
	resolverName := fmt.Sprintf("%sResolver", name)

	//
	// Generate resolver interface
	//
	// ... comment: Describe resolver interface and usage
	// ... method:  Serialize
	// ... method:  ParseValue
	// ... method:  ParseLiteral
	//

	code.Commentf(`//
// %s represents a collection of methods whose products represent the input and
// response values of a scalar type.
//
//  == Example input SDL
//
//    """
//    Timestamps are great.
//    """
//    scalar Timestamp
//
//  == Example generated interface
//
//    // DateResolver ...
//    type DateResolver interface {
//      // Serialize an internal value to include in a response.
//      Serialize(interface{}) interface{}
//      // ParseValue parses an externally provided value to use as an input.
//      ParseValue(interface{}) interface{}
//      // ParseLiteral parses an externally provided literal value to use as an input.
//      ParseLiteral(ast.Value) interface{}
//    }
//
//  == Example implementation
//
//    // MyDateResolver implements DateResolver interface
//    type MyDateResolver struct {
//      defaultTZ *time.Location
//      logger    logrus.LogEntry
//    }
//
//    // Serialize serializes given date into RFC 943 compatible string.
//    func (r *MyDateResolver) Serialize(val interface{}) interface{} {
//      // ... implementation details ...
//    }
//
//    // ParseValue takes given value and coerces it into an instance of Time.
//    func (r *MyDateResolver) ParseValue(val interface{}) interface{} {
//      // ... implementation details ...
//      // eg. if val is an int use time.At(), if string time.Parse(), etc.
//    }
//
//    // ParseValue takes given value and coerces it into an instance of Time.
//    func (r *MyDateResolver) ParseValue(val ast.Value) interface{} {
//      // ... implementation details ...
//      //
//      // eg.
//      //
//      // if string value return value,
//      // if IntValue Atoi and return value,
//      // etc.
//    }`,
		resolverName,
	)
	// Generate resolver interface.
	code.Type().Id(resolverName).Interface(
		// Serialize method.
		jen.Comment("Serialize an internal value to include in a response."),
		jen.Id("Serialize").Params(jen.Id("interface{}")).Interface(),

		// ParseValue method.
		jen.Comment("ParseValue parses an externally provided value to use as an input."),
		jen.Id("ParseValue").Params(jen.Id("interface{}")).Interface(),

		// ParseLiteral method.
		jen.Comment("ParseLiteral parses an externally provided literal value to use as an input."),
		jen.Id("ParseLiteral").Params(jen.Qual(astPkg, "Value")).Interface(),
	)

	//
	// Generate type definition
	//
	// ... comment: Include description in comment
	// ... panic callbacks panic if not configured
	//

	// Scalar description
	typeDesc := fetchDescription(node)

	// To appease the linter ensure that the the description of the scalar begins
	// with the name of the resulting method.
	desc := typeDesc
	if hasPrefix := strings.HasPrefix(typeDesc, name); !hasPrefix {
		desc = name + " " + desc
	}

	//
	// Generates thunk that returns new instance of scalar config
	//
	//  == Example input SDL
	//
	//    """
	//    Timestamps are great.
	//    """
	//    scalar Timestamp
	//
	//  == Example output
	//
	//   // Timestamps are great
	//   func Timestamp() graphql.ScalarConfig {
	//     return graphql.ScalarConfig{
	//       Name:         "Timestamp",
	//       Description:  "Timestamps are great.",
	//       Serialize:    // ...
	//       ParseValue:   // ...
	//       ParseLiteral: // ...
	//     }
	//   }
	//
	code.Comment(desc)
	code.Func().Id(name).Params().Qual(graphqlGoPkg, "ScalarConfig").Block(
		jen.Return(jen.Qual(graphqlGoPkg, "ScalarConfig").Values(jen.Dict{
			// Name & description
			jen.Id("Name"):        jen.Lit(name),
			jen.Id("Description"): jen.Lit(typeDesc),

			// Resolver funcs
			jen.Id("Serialize"): jen.Func().Params(jen.Id("_").Interface()).Interface().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
			),
			jen.Id("ParseValue"): jen.Func().Params(jen.Id("_").Interface()).Interface().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
			),
			jen.Id("ParseLiteral"): jen.Func().Params(jen.Id("_").Qual(astPkg, "Value")).Interface().Block(
				jen.Comment(missingResolverNote),
				jen.Panic(jen.Lit("Unimplemented; see "+resolverName+".")),
			),
		})),
	)

	return code
}
