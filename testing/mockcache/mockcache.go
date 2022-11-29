package mockcache

import (
	corev2 "github.com/sensu/core/v2"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
	"github.com/stretchr/testify/mock"
)

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(namespace string) []cachev2.Value[*corev2.Silenced, corev2.Silenced] {
	args := m.Called(namespace)
	return args.Get(0).([]cachev2.Value[*corev2.Silenced, corev2.Silenced])
}
