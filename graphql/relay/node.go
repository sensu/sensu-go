package relay

import (
	"github.com/Sirupsen/logrus"
	"github.com/graphql-go/graphql"
	"github.com/sensu/sensu-go/graphql/globalid"
	"golang.org/x/net/context"
)

// A NodeResolver describes an object that contains a globally unique ID.
type NodeResolver struct {
	Object     *graphql.Object
	Translator globalid.Resource

	Resolve  func(context.Context, globalid.Components) (interface{}, error)
	IsKindOf func(globalid.Components) bool
}

// NodeRegister stores list of node resolvers
type NodeRegister map[string][]NodeResolver

// RegisterResolver registers a new node resolver and adds it to the register
func (register NodeRegister) RegisterResolver(resolver NodeResolver) {
	translatorName := resolver.Translator.Name()
	entry := register[translatorName]
	entry = append(entry, resolver)

	if len(entry) > 1 && entry[0].IsKindOf == nil && entry[1].IsKindOf == nil {
		logEntry := logger.WithFields(logrus.Fields{
			"resolver": resolver,
			"globalid": translatorName,
		})
		logEntry.Fatal(
			"Failed to register resolver because you already have one resolver " +
				"registered to this globalid. If you do intend to register more than " +
				"on resolver you must set the IsKindOf field, so that the Register " +
				"can properly route fetches.",
		)
	}

	register[translatorName] = entry
}

// Lookup given parsed globalid return valid resolver.
func (register NodeRegister) Lookup(components globalid.Components) *NodeResolver {
	entries := register[components.Resource()]
	entriesLen := len(entries)

	logEntry := logger.WithField("registerContents", register)
	logEntry.Info("lookup")

	if entriesLen > 1 {
		for _, entry := range entries {
			if entry.IsKindOf(components) {
				return &entry
			}
		}
	} else if entriesLen == 1 {
		return &entries[0]
	}

	return nil
}
