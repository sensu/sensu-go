package globalid

//
// Translators
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

// A Translator represents something that is globally identifable
type Translator interface {
	ForResourceNamed() string
	Encoder
	Decoder
}
