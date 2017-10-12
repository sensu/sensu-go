package globalid

//
// Setup global instance of register
//

// Register & friends
var register = NewRegister()
var registrar = NewRegistrar(register)
var registerResource = registrar.Add

// Lookup given ID components return applicable encoder
var Lookup = register.Lookup

// ReverseLookup finds Encoder capable of encoding record as global ID
var ReverseLookup = register.ReverseLookup
