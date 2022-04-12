package mockcache

import (
	"github.com/sensu/sensu-go/backend/store/cache"
	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(namespace string) []cache.Value {
	args := m.Called(namespace)
	return args.Get(0).([]cache.Value)
}
