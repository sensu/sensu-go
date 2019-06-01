package schedulerd

import (
	"encoding/json"
	"fmt"

	time "github.com/echlebek/timeproxy"
	"github.com/robfig/cron"
	"github.com/sensu/sensu-go/agent"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
)

// matchEntities matches the provided list of entities to the entity attributes
// configured in the proxy request
func matchEntities(entities []*types.Entity, proxyRequest *types.ProxyRequests) []*types.Entity {
	matched := make([]*types.Entity, 0, len(entities))
	synthesizedEntities := make([]interface{}, 0, len(entities))
	for _, entity := range entities {
		synth := dynamic.Synthesize(entity)
		synthesizedEntities = append(synthesizedEntities, synth)
	}

	results, err := js.MatchEntities(proxyRequest.EntityAttributes, synthesizedEntities)
	if err != nil {
		logger.Error(fmt.Errorf("error evaluating proxy entities: %s", err))
		return nil
	}

	if got, want := len(results), len(entities); got != want {
		logger.Error(fmt.Errorf("mismatched result and entity lengths: (%d != %d)", got, want))
		return nil
	}

	for i, result := range results {
		if result {
			matched = append(matched, entities[i])
		}
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

// calculateSplayInterval calculates the duration between publishing proxy
// requests to each individual entity (based on a configurable splay %)
func calculateSplayInterval(check *types.CheckConfig, numEntities int) (time.Duration, error) {
	next := time.Second * time.Duration(check.Interval)
	if check.Cron != "" {
		schedule, err := cron.ParseStandard(check.Cron)
		if err != nil {
			return 0, err
		}
		now := time.Now()
		then := schedule.Next(now)
		next = then.Sub(now)
		if next < 5*time.Second {
			now = time.Now().Add(next + time.Second)
			then = schedule.Next(now)
			next = then.Sub(now)
		}
	}
	splayCoverage := float64(check.ProxyRequests.SplayCoverage)
	if splayCoverage == 0 {
		splayCoverage = types.DefaultSplayCoverage
	}
	timeSlice := splayCoverage / 100.0 / float64(numEntities)
	splay := time.Duration(float64(next) * timeSlice)
	return splay, nil
}
