package testing

import "github.com/stretchr/testify/mock"

// MockClient uses mock package to allow your tests to easily
// mock out the results of any SensuClient operation.
type MockClient struct {
	mock.Mock
}
