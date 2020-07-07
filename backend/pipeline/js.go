package pipeline

import (
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/js"
)

type FilterExecutionEnvironment struct {
	// Funcs are a list of named functions to be supplied to the JS environment.
	Funcs map[string]interface{}

	// Assets are a set of javascript assets to be evaluated before filter
	// execution.
	Assets asset.RuntimeAssetSet

	// Event is the Sensu event to be supplied to the filter execution environment.
	Event interface{}
}

func (f *FilterExecutionEnvironment) Eval(expression string) (bool, error) {
	parameters := map[string]interface{}{}
	var assets asset.RuntimeAssetSet
	if f != nil {
		parameters["event"] = f.Event
		for k, v := range f.Funcs {
			parameters[k] = v
		}
		assets = f.Assets
	}
	return js.Evaluate(expression, parameters, assets)
}
