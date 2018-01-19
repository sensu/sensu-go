package schedulerd

import (
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/util/eval"
)

// matchEntities matches the provided list of entities to the entity attributes
// configured in the proxy request
func matchEntities(entities []*types.Entity, proxyRequest *types.ProxyRequests) []*types.Entity {
	matched := []*types.Entity{}

OUTER:
	for _, entity := range entities {
		for _, expression := range proxyRequest.EntityAttributes {
			parameters := map[string]interface{}{"entity": entity}

			result, err := eval.Evaluate(expression, parameters)
			if err != nil {
				// Skip to the next entity
				logger.WithError(err).Errorf("expression '%s' is invalid", expression)
				continue OUTER
			}

			// Check if the expression returned a negative result, and if so, skip to
			// the next entity
			if !result {
				continue OUTER
			}
		}

		matched = append(matched, entity)
	}

	return matched
}
