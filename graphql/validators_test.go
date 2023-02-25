package graphql

import (
	"testing"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/visitor"
)

var (
	schema   *graphql.Schema
	typeInfo *graphql.TypeInfo
)

// init stubbed GraphQL schema
func init() {
	// NOTE: the max-depth validator doesn't care about a valid schema, only depth of nodes in the graph
	queryType := graphql.NewObject(graphql.ObjectConfig{
		Name: "Query",
		Fields: graphql.Fields{
			"foobar": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{Query: queryType})
	if err != nil {
		panic(err)
	}
	typeInfo = graphql.NewTypeInfo(&graphql.TypeInfoConfig{Schema: &schema})
}

// helper to parse a graphql query document
func parseQuery(t *testing.T, q string) *ast.Document {
	t.Helper()
	astDoc, err := parser.Parse(parser.ParseParams{Source: q})
	if err != nil {
		t.Fatalf("parse failed: %s", err)
	}
	return astDoc
}

// helper to assert query depth matches expected
func testDepth(t *testing.T, query string, depthLimit int, expectedDepth int) *maxDepthRule {
	t.Helper()
	astDoc := parseQuery(t, query)
	context := graphql.NewValidationContext(schema, astDoc, typeInfo)
	rule := newMaxDepthRule(depthLimit)
	rule.context = context
	visitor.Visit(astDoc, rule.maxDepthVisitorOptions(), nil)
	if rule.depth != expectedDepth {
		t.Fatalf("wrong depth expected: want=%d got=%d", expectedDepth, rule.depth)
	}
	return rule
}

// helper to assert errors set when depth limit reached
func testMaxDepthError(t *testing.T, query string, depthLimit int) *maxDepthRule {
	t.Helper()
	astDoc := parseQuery(t, query)
	context := graphql.NewValidationContext(schema, astDoc, typeInfo)
	rule := newMaxDepthRule(depthLimit)
	rule.context = context
	visitor.Visit(astDoc, rule.maxDepthVisitorOptions(), nil)
	if len(context.Errors()) == 0 {
		t.Fatalf("expected errors but none returned")
	}
	return rule
}

func TestMaxDepth(t *testing.T) {
	testDepth(t, `query MyQuery {   # depth 0
		postgres {                    # depth 1
			healthy				              # depth 2
		}
	}`, 5, 2)
}

func TestMaxDepthFragment(t *testing.T) {
	testDepth(t, `
	query MyQuery {   # depth 0
		pet {           # depth 1
			breed				  # depth 2
			...dog 				# depth 2
		}
	}
	fragment dog on Pet {
		name           # depth 2
		owner {
			name         # depth 3
		}
	}`, 5, 3)
}

// test don't break on cyclical fragment spread
func TestMaxDepthFragmentCycle(t *testing.T) {
	testDepth(t, `
		fragment X on Query { ...Y }
		fragment Y on Query { ...X }
		query {
			...X
		}
	`, 5, 1)
}

func TestMaxDepthInlineFragment(t *testing.T) {
	testDepth(t, `
	query MyQuery {     # depth 0
		pet {             # depth 1
			breed				    # depth 2
			... on Pet {    # depth 2
				name          # depth 2
			} 
		}
	}`, 5, 2)
}

func TestMaxDepthFail(t *testing.T) {
	testMaxDepthError(t, `query MyQuery {    # depth 0
		postgres {                             # depth 1
			healthy				                       # depth 2 <-- max depth reached here
		}
	}`, 1)
}
