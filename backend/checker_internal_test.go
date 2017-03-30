package backend

import (
	"errors"
	"sync"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStore struct {
	mock.Mock
	// Entities
	getEntityByID func(id string) (*types.Entity, error)
	updateEntity  func(e *types.Entity) error
	deleteEntity  func(e *types.Entity) error
	getEntities   func() ([]*types.Entity, error)

	// Checks
	getChecks         func() ([]*types.Check, error)
	getCheckByName    func(name string) (*types.Check, error)
	deleteCheckByName func(name string) error
	updateCheck       func(check *types.Check) error
}

func (m *mockStore) GetEntityByID(id string) (*types.Entity, error) {
	return m.getEntityByID(id)
}

func (m *mockStore) UpdateEntity(e *types.Entity) error {
	return m.updateEntity(e)
}

func (m *mockStore) DeleteEntity(e *types.Entity) error {
	return m.deleteEntity(e)
}

func (m *mockStore) GetEntities() ([]*types.Entity, error) {
	return m.getEntities()
}

func (m *mockStore) GetChecks() ([]*types.Check, error) {
	return m.getChecks()
}

func (m *mockStore) GetCheckByName(name string) (*types.Check, error) {
	return m.getCheckByName(name)
}

func (m *mockStore) DeleteCheckByName(name string) error {
	return m.deleteCheckByName(name)
}

func (m *mockStore) UpdateCheck(check *types.Check) error {
	return m.updateCheck(check)
}

func TestCheckerReconcile(t *testing.T) {
	st := &mockStore{}
	st.getChecks = func() ([]*types.Check, error) {
		checks := []*types.Check{
			{
				Name:          "check1",
				Interval:      60,
				Command:       "command",
				Subscriptions: []string{},
			},
		}
		return checks, nil
	}
	st.getCheckByName = func(name string) (*types.Check, error) {
		return nil, nil
	}

	c := &Checker{
		Store:           st,
		schedulersMutex: &sync.Mutex{},
		schedulers:      map[string]*CheckScheduler{},
	}
	c.reconcile()

	assert.Equal(t, 1, len(c.schedulers))

	st.getChecks = func() ([]*types.Check, error) {
		return nil, errors.New("")
	}

	assert.Error(t, c.reconcile())
}
