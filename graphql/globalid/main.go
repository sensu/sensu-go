package globalid

//
// Setup global instance of register
//

// Register & friends
var register = NewRegister()
var registrar = NewRegistrar(register)
var registerTranslator = registrar.Add

// Lookup given ID components return applicable encoder
var Lookup = register.Lookup

// ReverseLookup finds Encoder capable of encoding record as global ID
var ReverseLookup = register.ReverseLookup

// Decoder parse, look up decoder & decode.
func Decode(gid string) (Components, error) {
	standardComponents, err := Parse(gid)
	if err != nil {
		return standardComponents, err
	}

	decoder, err := Lookup(standardComponents)
	if err != nil {
		return standardComponents, err
	}

	components := decoder.Decode(standardComponents)
	return components, nil
}
