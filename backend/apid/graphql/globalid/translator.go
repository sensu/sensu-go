package globalid

import "context"

//
// Translators
//

// An Encoder can encode global IDs for a specific resource
type Encoder interface {
	IsResponsible(interface{}) bool
	Encode(context.Context, interface{}) Components
	EncodeToString(context.Context, interface{}) string
}

// A Decoder can decode global IDs for a specific resource
type Decoder interface {
	Decode(StandardComponents) Components
}

// A Translator represents something that is globally identifable
type Translator interface {
	ForResourceNamed() string
	Encoder
	Decoder
}
