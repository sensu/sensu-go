package globalid

import (
	"context"
	"reflect"

	"github.com/sensu/sensu-go/types"
)

//
// Standard ID Components
//

// NamedComponents adds Name method that returns first element of
// global ID's uniqueComponents.
type NamedComponents struct{ *StandardComponents }

// Name method returns first element of global ID's uniqueComponents. Congurent
// with most Sensu records Check#Name, Entity#ID, Asset#name, etc.
func (n NamedComponents) Name() string {
	return n.uniqueComponent
}

//
// Standard Translator
//

type encoderFunc func(context.Context, interface{}) Components
type decoderFunc func(StandardComponents) Components

type commonTranslator struct {
	name              string
	isResponsibleFunc func(interface{}) bool
	encodeFunc        encoderFunc
	decodeFunc        decoderFunc
}

func (r commonTranslator) ForResourceNamed() string {
	return r.name
}

func (r commonTranslator) IsResponsible(record interface{}) bool {
	return r.isResponsibleFunc(record)
}

func (r commonTranslator) Encode(ctx context.Context, record interface{}) Components {
	components := r.encodeFunc(ctx, record)
	return components
}

func (r commonTranslator) EncodeToString(ctx context.Context, record interface{}) string {
	components := r.Encode(ctx, record)
	return components.String()
}

func (r commonTranslator) Decode(components StandardComponents) Components {
	return r.decodeFunc(components)
}

//
// Helpers
//

func addMultitenantFields(c *StandardComponents, r types.MultitenantResource) {
	c.namespace = r.GetNamespace()
}

// standardDecoder instantiates new NamedComponents composite.
func standardDecoder(components StandardComponents) Components {
	return NamedComponents{&components}
}

// standardEncoder encodes record given name and unique field name
func standardEncoder(name string, fNames ...string) encoderFunc {
	return func(ctx context.Context, record interface{}) Components {
		// Retrieve the value of the specified field
		fVal := reflect.ValueOf(record)
		for _, fName := range fNames {
			fVal = reflect.Indirect(fVal)
			fVal = fVal.FieldByName(fName)
		}

		// Add string value of field to global id components
		components := Encode(ctx, record)
		components.resource = name
		components.uniqueComponent = fVal.String()

		// Add namespace to global id components
		if multiRecord, ok := record.(types.MultitenantResource); ok {
			addMultitenantFields(components, multiRecord)
		}

		return components
	}
}
