package mockstore

import "github.com/stretchr/testify/mock"

// MockStore is a store used for testing. When using the MockStore in unit
// tests, stub out the behavior you wish to test against by assigning the
// appropriate function to the appropriate Func field. If you have forgotten
// to stub a particular function, the program will panic.
type MockStore struct {
	mock.Mock
}
