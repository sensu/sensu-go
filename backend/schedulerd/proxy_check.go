package schedulerd

import (
	"encoding/json"
	"fmt"

	time "github.com/echlebek/timeproxy"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/sirupsen/logrus"
)

// matchEntities matches the provided list of entities to the entity attributes
// configured in the proxy request
func matchEntities(entities []*types.Entity, proxyRequest *types.ProxyRequests) []*types.Entity {
	matched := []*types.Entity{}

OUTER:
	for _, entity := range entities {
		synth := dynamic.Synthesize(entity)
		for _, expression := range proxyRequest.EntityAttributes {

			parameters := map[string]interface{}{"entity": synth}

			result, err := js.Evaluate(expression, parameters, nil)
			if err != nil {
				fields := logrus.Fields{
					"expression": expression,
					"entity":     entity.Name,
					"namespace":  entity.Namespace,
				}
				// Only report an error if the expression's syntax is invalid
				switch err.(type) {
				case js.SyntaxError:
					logger.WithFields(fields).WithError(err).Error("syntax error")
				default:
					logger.WithFields(fields).WithError(err).Debug("skipping expression")
				}
				// Skip to the next entity
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

// substituteProxyEntityTokens substitutes entity tokens in the proxy check definition. If
// there are unmatched entity tokens, it returns an error.
func substituteProxyEntityTokens(entity *types.Entity, check *types.CheckConfig) (*types.CheckConfig, error) {
	// Extract the extended attributes from the entity and combine them at the
	// top-level so they can be easily accessed using token substitution
	synthesizedEntity := dynamic.Synthesize(entity)

	// Substitute tokens within the check configuration with the synthesized
	// entity
	checkBytes, err := agent.TokenSubstitution(synthesizedEntity, check)
	if err != nil {
		return nil, err
	}

	substitutedCheck := &types.CheckConfig{}

	// Unmarshal the check configuration obtained after the token substitution
	// back into the check config struct
	err = json.Unmarshal(checkBytes, substitutedCheck)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal the check: %s", err)
	}

	substitutedCheck.ProxyEntityName = entity.Name
	return substitutedCheck, nil
}

// calculateSplayInterval calculates how many seconds between publishing proxy
// requests to each individual entity (based on a configurable splay %)
func calculateSplayInterval(check *types.CheckConfig, numEntities float64) (float64, error) {
	var err error
	next := time.Duration(time.Second * time.Duration(check.Interval))
	if check.Cron != "" {
		if next, err = NextCronTime(time.Now(), check.Cron); err != nil {
			return 0, err
		}
	}
	splayCoverage := float64(check.ProxyRequests.SplayCoverage)
	if splayCoverage == 0 {
		splayCoverage = types.DefaultSplayCoverage
	}
	splay := next.Seconds() * (splayCoverage / 100.0) / numEntities
	return splay, nil
}
