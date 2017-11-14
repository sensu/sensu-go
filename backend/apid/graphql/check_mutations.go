package graphqlschema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/mitchellh/mapstructure"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
	"golang.org/x/net/context"
)

var createCheckMutation *graphql.Field
var updateCheckMutation *graphql.Field
var destroyCheckMutation *graphql.Field

func init() {
	initCheckConfigType()
	initDestroyCheckMutation()
	initUpdateCheckMutation()
	initCreateCheckMutation()
}

func initCreateCheckMutation() {
	createCheckMutation = relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "CreateCheck",
		InputFields: graphql.InputObjectConfigFieldMap{
			"name":              NewInputFromObjectField(checkConfigType, "name", nil),
			"organization":      NewInputFromObjectField(checkConfigType, "organization", "default"),
			"environment":       NewInputFromObjectField(checkConfigType, "environment", "default"),
			"interval":          NewInputFromObjectField(checkConfigType, "interval", nil),
			"subscriptions":     NewInputFromObjectField(checkConfigType, "subscriptions", nil),
			"command":           NewInputFromObjectField(checkConfigType, "command", nil),
			"handlers":          NewInputFromObjectField(checkConfigType, "handlerNames", nil),
			"highFlapThreshold": NewInputFromObjectField(checkConfigType, "highFlapThreshold", nil),
			"lowFlapThreshold":  NewInputFromObjectField(checkConfigType, "lowFlapThreshold", nil),
			"publish":           NewInputFromObjectField(checkConfigType, "publish", true),
			"runtimeAssets":     NewInputFromObjectField(checkConfigType, "runtimeAssetNames", nil),
		},
		OutputFields: graphql.Fields{
			"check": &graphql.Field{Type: checkConfigType},
		},
		MutateAndGetPayload: func(inputs map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			var check = types.CheckConfig{}
			var results = map[string]interface{}{}

			if err := mapstructure.Decode(inputs, &check); err != nil {
				logger.WithField("inputs", inputs).WithError(err).Error("unable to decode input")
				return results, err
			}

			store := ctx.Value(types.StoreKey).(store.Store)
			controller := actions.NewCheckController(store)

			if err := controller.Create(ctx, check); err != nil {
				logger.WithField("inputs", inputs).WithError(err).Debug("unable to create check")
				return results, err
			}

			results["check"] = &check
			return results, nil
		},
	})
}

func initUpdateCheckMutation() {
	updateCheckMutation = relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "UpdateCheck",
		InputFields: graphql.InputObjectConfigFieldMap{
			"name":              NewInputFromObjectField(checkConfigType, "name", nil),
			"organization":      NewInputFromObjectField(checkConfigType, "organization", "default"),
			"environment":       NewInputFromObjectField(checkConfigType, "environment", "default"),
			"interval":          NewInputFromObjectField(checkConfigType, "interval", nil),
			"subscriptions":     NewInputFromObjectField(checkConfigType, "subscriptions", nil),
			"command":           NewInputFromObjectField(checkConfigType, "command", nil),
			"handlers":          NewInputFromObjectField(checkConfigType, "handlerNames", nil),
			"highFlapThreshold": NewInputFromObjectField(checkConfigType, "highFlapThreshold", nil),
			"lowFlapThreshold":  NewInputFromObjectField(checkConfigType, "lowFlapThreshold", nil),
			"publish":           NewInputFromObjectField(checkConfigType, "publish", true),
			"runtimeAssets":     NewInputFromObjectField(checkConfigType, "runtimeAssetNames", nil),
		},
		OutputFields: graphql.Fields{
			"check": &graphql.Field{Type: checkConfigType},
		},
		MutateAndGetPayload: func(inputs map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			var check = types.CheckConfig{}
			var results = map[string]interface{}{}

			if err := mapstructure.Decode(inputs, &check); err != nil {
				logger.WithField("inputs", inputs).WithError(err).Error("unable to decode input")
				return results, err
			}

			store := ctx.Value(types.StoreKey).(store.Store)
			controller := actions.NewCheckController(store)

			if err := controller.Update(ctx, check); err != nil {
				logger.WithField("inputs", inputs).WithError(err).Debug("unable to update check")
				return results, err
			}

			results["check"] = &check
			return results, nil
		},
	})
}

func initDestroyCheckMutation() {
	destroyCheckMutation = relay.MutationWithClientMutationID(relay.MutationConfig{
		Name: "DestroyCheck",
		InputFields: graphql.InputObjectConfigFieldMap{
			"id": NewInputFromObjectField(checkConfigType, "id", nil),
		},
		OutputFields: graphql.Fields{
			"destroyedCheckId": &graphql.Field{
				Description: "The ID of the deleted check",
				Type:        graphql.NewNonNull(graphql.ID),
			},
		},
		MutateAndGetPayload: func(inputs map[string]interface{}, info graphql.ResolveInfo, ctx context.Context) (map[string]interface{}, error) {
			var results = map[string]interface{}{}

			// TODO (JK): handle the error from DecodeIDFromInputs()
			components, _ := DecodeIDFromInputs(inputs, "id")
			params := actions.QueryParams{"id": components.UniqueComponent()}
			ctx = SetContextFromComponents(ctx, components)

			store := ctx.Value(types.StoreKey).(store.Store)
			controller := actions.NewCheckController(store)

			if err := controller.Destroy(ctx, params); err != nil {
				logger.WithField("inputs", inputs).WithError(err).Debug("unable to delete check")
				return results, err
			}

			results["destroyedCheckId"] = components.String()
			return results, nil
		},
	})
}
