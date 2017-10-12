package graphqlschema

import (
	"regexp"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/graphql-go/graphql"
	"golang.org/x/net/context"
)

var matchWhitespaceRegex = regexp.MustCompile(`\s+`)

// Execute given query against Sensu schema
func Execute(ctx context.Context, q string, vars *map[string]interface{}) *graphql.Result {
	exec := NewExecution(ctx, q)

	if vars != nil {
		exec.SetVariables(*vars)
	}

	return exec.Run()
}

// Execution runs given query against Sensu schema.
type Execution struct {
	Query     string
	Context   context.Context
	Variables map[string]interface{}
}

// NewExecution returns new Execution from given query string
func NewExecution(ctx context.Context, q string) Execution {
	e := Execution{}
	e.Query = q
	e.Context = ctx

	return e
}

// SetVariables sets execution variables given map
func (e *Execution) SetVariables(vars map[string]interface{}) {
	e.Variables = vars
}

// Run executes given query and parameters against Sensu schema.
func (e *Execution) Run() *graphql.Result {
	params := graphql.Params{
		Context:        e.Context,
		RequestString:  e.Query,
		Schema:         Schema,
		VariableValues: e.Variables,
	}

	if logger.Logger.Level >= logrus.DebugLevel {
		formattedQuery := e.Query
		formattedQuery = strings.Replace(formattedQuery, "\n", " ", -1)
		formattedQuery = matchWhitespaceRegex.ReplaceAllLiteralString(formattedQuery, " ")
		logger.WithField("query", formattedQuery).Debug("executing GraphQL query")
	}

	return graphql.Do(params)
}
