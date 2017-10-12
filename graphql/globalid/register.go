package globalid

import (
	"fmt"

	"github.com/Sirupsen/logrus"
)

var (
	defaultLogger = logrus.WithField("component", "globalid")
)

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
func (registrar *Registrar) Add(resource Resource) {
	name := resource.Name()

	registrar.Logger.WithField("name", name).Debug("registering resource")
	registrar.register.resources[name] = resource
}

//
// Register
//

// A Register holds resource definitions for easy lookup
type Register struct {
	Logger    logrus.FieldLogger
	resources map[string]Resource
}

// NewRegister instatiates new register.
func NewRegister() Register {
	logger := defaultLogger.WithField("subcomponent", "register")
	return Register{Logger: logger}
}

// Lookup given ID components return applicable encoder
func (r Register) Lookup(components StandardComponents) (Decoder, error) {
	entry := r.Logger.WithField("resource", components.Resource())
	entry.Debug("looking up decoder")

	if resource, ok := r.resources[components.Resource()]; ok {
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
	for _, encoder := range r.resources {
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

//
// Resources
//

// An Encoder can encode global IDs for a specific resource
type Encoder interface {
	IsResponsible(interface{}) bool
	Encode(interface{}) Components
	EncodeToString(interface{}) string
}

// A Decoder can decode global IDs for a specific resource
type Decoder interface {
	Decode(StandardComponents) Components
}

// A Resource represents something that is globally identifable
type Resource interface {
	Name() string
	Encoder
	Decoder
}
