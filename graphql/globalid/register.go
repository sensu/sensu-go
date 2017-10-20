package globalid

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

var defaultLogger = logrus.WithField("component", "graphql/globalid")

//
// Registrar
//

// A Registrar is used to register new resources
type Registrar struct {
	Logger   logrus.FieldLogger
	register Register
}

// NewRegistrar instatiates new registrar.
func NewRegistrar(register Register) Registrar {
	logger := defaultLogger.WithField("subcomponent", "registrar")

	return Registrar{
		Logger:   logger,
		register: register,
	}
}

// Add resource to global ID register
func (registrar *Registrar) Add(translator Translator) {
	name := translator.ForResourceNamed()

	registrar.Logger.WithField("name", name).Debug("registering resource")
	registrar.register.translators[name] = translator
}

//
// Register
//

// A Register holds resource definitions for easy lookup
type Register struct {
	Logger      logrus.FieldLogger
	translators map[string]Translator
}

// NewRegister instatiates new register.
func NewRegister() Register {
	logger := defaultLogger.WithField("subcomponent", "register")

	return Register{
		Logger:      logger,
		translators: make(map[string]Translator),
	}
}

// Lookup given ID components return applicable encoder
func (r Register) Lookup(components StandardComponents) (Decoder, error) {
	entry := r.Logger.WithField("resource", components.Resource())
	entry.Debug("looking up decoder")

	if resource, ok := r.translators[components.Resource()]; ok {
		return resource, nil
	}

	return nil, fmt.Errorf(
		"global ID decoder could not be found for '%s'",
		components.Resource(),
	)
}

// ReverseLookup finds Encoder capable of encoding record as global ID
func (r Register) ReverseLookup(record interface{}) (Encoder, error) {
	recordType := fmt.Sprintf("%T", record)
	entry := r.Logger.WithField("record", recordType)
	entry.Debug("looking up encoder")

	// iterate through our resources until we find one the one that can encode the
	// given record.
	for _, encoder := range r.translators {
		if encoder.IsResponsible(record) {
			return encoder, nil
		}
	}

	// if not found try to be as helpful as humanly possible
	return nil, fmt.Errorf(
		"global ID encoder could not be found for '%s'",
		recordType,
	)
}
