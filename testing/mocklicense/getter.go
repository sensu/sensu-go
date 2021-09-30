package mocklicense

import (
	"github.com/stretchr/testify/mock"
)

// Getter ...
type Getter struct {
	mock.Mock
}

// Get ...
func (m *Getter) Get() string {
	args := m.Called()
	return args.Get(0).(string)
}
