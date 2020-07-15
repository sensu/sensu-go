package pipeline

import (
	"context"

	"github.com/robertkrimen/otto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types/dynamic"
)

const pipelineRoleName = "system:pipeline"

// PipelineFilterFuncs gets patched by enterprise sensu-go
var PipelineFilterFuncs map[string]interface{}

type FilterExecutionEnvironment struct {
	// Funcs are a list of named functions to be supplied to the JS environment.
	Funcs map[string]interface{}

	// Assets are a set of javascript assets to be evaluated before filter
	// execution.
	Assets asset.RuntimeAssetSet

	// Event is the Sensu event to be supplied to the filter execution environment.
	Event interface{}
}

func (f *FilterExecutionEnvironment) Eval(ctx context.Context, expression string) (bool, error) {
	// these claims set up authorization credentials for the event client.
	// the actual roles and role bindings are generated when namespaces are
	// created, for the purposes of the system.
	claims := &corev2.Claims{
		Groups: []string{pipelineRoleName},
		Provider: corev2.AuthProviderClaims{
			ProviderID:   "system",
			ProviderType: "basic",
			UserID:       pipelineRoleName,
		},
	}
	ctx = context.WithValue(ctx, corev2.ClaimsKey, claims)
	parameters := map[string]interface{}{}
	var assets asset.RuntimeAssetSet
	if f != nil {
		parameters["event"] = f.Event
		assets = f.Assets
	}
	var result bool
	err := js.WithOttoVM(assets, func(vm *otto.Otto) (err error) {
		if f != nil {
			funcs := make(map[string]interface{}, len(f.Funcs))
			for k, v := range f.Funcs {
				funcs[k] = dynamic.Function(ctx, vm, v)
			}
			parameters["sensu"] = funcs
		}
		result, err = js.EvalPredicateWithVM(vm, parameters, expression)
		return err
	})
	return result, err
}
