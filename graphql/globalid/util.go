package globalid

import (
	"reflect"

	"github.com/sensu/sensu-go/types"
)

//
// Standard ID Components
//

// NamedComponents adds Name method that returns first element of
// global ID's uniqueComponents.
type NamedComponents struct{ StandardComponents }

// Name method returns first element of global ID's uniqueComponents. Congurent
// with most Sensu records Check#Name, Entity#ID, Asset#name, etc.
func (n NamedComponents) Name() string {
	return n.uniqueComponents[0]
}

//
// Standard Resource
//

type encoderFunc func(interface{}) Components
type decoderFunc func(StandardComponents) Components

type commonResource struct {
	name              string
	isResponsibleFunc func(interface{}) bool
	encodeFunc        encoderFunc
	decodeFunc        decoderFunc
}

func (r commonResource) Name() string {
	return r.name
}

func (r commonResource) IsResponsible(record interface{}) bool {
	return r.isResponsibleFunc(record)
}

func (r commonResource) Encode(record interface{}) Components {
	components := r.encodeFunc(record)
	return components
}

func (r commonResource) EncodeToString(record interface{}) string {
	components := r.Encode(record)
	return components.String()
}

func (r commonResource) Decode(components StandardComponents) Components {
	return r.decodeFunc(components)
}

//
// Helpers
//

func addMultitenantFields(c *StandardComponents, r types.MultitenantResource) {
	c.organization = r.GetOrg()
	c.environment = r.GetEnv()
}

// newComponentsWith returns new instance of StandardComponents w/ name and ids
func newComponentsWith(resourceName string, uids ...string) StandardComponents {
	return StandardComponents{
		resource:         resourceName,
		uniqueComponents: uids,
	}
}

// standardDecoder instantiates new NamedComponents composite.
func standardDecoder(components StandardComponents) Components {
	return NamedComponents{components}
}

// standardEncoder encodes record given name and unique field name
func standardEncoder(name string, fNames ...string) encoderFunc {
	return func(record interface{}) Components {
		// Retrieve the value of the specified field
		fVal := reflect.ValueOf(record)
		for _, fName := range fNames {
			fVal = reflect.Indirect(fVal)
			fVal = fVal.FieldByName(fName)
		}

		// Add string value of field to global id components
		components := newComponentsWith(name, fVal.String())

		// Add org & env to global id components
		if multiRecord, ok := record.(types.MultitenantResource); ok {
			addMultitenantFields(&components, multiRecord)
		}

		return components
	}
}
