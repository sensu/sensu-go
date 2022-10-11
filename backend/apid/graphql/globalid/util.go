package globalid

import (
	"context"
	"reflect"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/types"
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

func addMultitenantFields(c *StandardComponents, r corev2.MultitenantResource) {
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
		if multiRecord, ok := record.(corev2.MultitenantResource); ok {
			addMultitenantFields(components, multiRecord)
		}

		return components
	}
}

type tmGetter interface {
	GetTypeMeta() corev2.TypeMeta
}

type namedResource interface {
	RBACName() string
}

// NewGenericTranslator creates a compatible translator for a given corev2 or
// corev3 resource.
func NewGenericTranslator(kind namedResource, name string) Translator {
	return &genericTranslator{kind: kind, name: name}
}

type genericTranslator struct {
	kind     interface{}
	name     string
	_kindVal *reflect.Value
}

func (g *genericTranslator) kindVal() *reflect.Value {
	if g._kindVal == nil {
		val := reflect.ValueOf(g.kind)
		g._kindVal = &val
	}
	return g._kindVal
}

// Returns the rbac name for the given resource
func (g *genericTranslator) ForResourceNamed() string {
	if g.name != "" {
		return g.name
	}
	tm := corev2.TypeMeta{}
	if getter, ok := g.kind.(tmGetter); ok {
		tm = getter.GetTypeMeta()
	} else {
		typ := reflect.Indirect(reflect.ValueOf(g.kind)).Type()
		tm = corev2.TypeMeta{
			Type:       typ.Name(),
			APIVersion: types.ApiVersion(typ.PkgPath()),
		}
	}
	g.name = tm.APIVersion + "." + tm.Type
	return g.name
}

// IsResponsible returns true if the given resource matches this translator
func (g *genericTranslator) IsResponsible(r interface{}) bool {
	return g.kindVal().Type().String() == reflect.ValueOf(r).Type().String()
}

// Encode produces id components for a given resource
func (g *genericTranslator) Encode(ctx context.Context, r interface{}) Components {
	name := g.ForResourceNamed()
	cmp := Encode(ctx, r)
	cmp.SetResource(name)
	return cmp
}

// EncodeToString returns a globalid for the given resource
func (g *genericTranslator) EncodeToString(ctx context.Context, r interface{}) string {
	return g.Encode(ctx, r).String()
}

// Decodes the given globalid into components
func (g *genericTranslator) Decode(cmp StandardComponents) Components {
	return &cmp
}
