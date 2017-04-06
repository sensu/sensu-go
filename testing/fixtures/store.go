package fixtures

import "github.com/sensu/sensu-go/types"

// FixtureStore implements the store.Store interface and stores all of the
// fixtures in memory.
type FixtureStore struct {
	Entities map[string]*types.Entity
	Handlers map[string]*types.Handler
	Mutators map[string]*types.Mutator
	Checks   map[string]*types.Check
}

// GetEntityByID ...
func (s *FixtureStore) GetEntityByID(id string) (*types.Entity, error) {
	e, ok := s.Entities[id]
	if !ok {
		return nil, nil
	}
	return e, nil
}

// UpdateEntity ...
func (s *FixtureStore) UpdateEntity(e *types.Entity) error {
	s.Entities[e.ID] = e
	return nil
}

// DeleteEntity ...
func (s *FixtureStore) DeleteEntity(e *types.Entity) error {
	delete(s.Entities, e.ID)
	return nil
}

// GetEntities ...
func (s *FixtureStore) GetEntities() ([]*types.Entity, error) {
	var entities []*types.Entity
	for _, v := range s.Entities {
		entities = append(entities, v)
	}
	return entities, nil
}

// Handlers ...

// GetHandlers ...
func (s *FixtureStore) GetHandlers() ([]*types.Handler, error) {
	var handlers []*types.Handler
	for _, v := range s.Handlers {
		handlers = append(handlers, v)
	}
	return handlers, nil
}

// GetHandlerByName ...
func (s *FixtureStore) GetHandlerByName(name string) (*types.Handler, error) {
	c, ok := s.Handlers[name]
	if !ok {
		return nil, nil
	}
	return c, nil
}

// DeleteHandlerByName ...
func (s *FixtureStore) DeleteHandlerByName(name string) error {
	delete(s.Handlers, name)
	return nil
}

// UpdateHandler ...
func (s *FixtureStore) UpdateHandler(handler *types.Handler) error {
	s.Handlers[handler.Name] = handler
	return nil
}

// Mutators ...

// GetMutators ...
func (s *FixtureStore) GetMutators() ([]*types.Mutator, error) {
	var mutators []*types.Mutator
	for _, v := range s.Mutators {
		mutators = append(mutators, v)
	}
	return mutators, nil
}

// GetMutatorByName ...
func (s *FixtureStore) GetMutatorByName(name string) (*types.Mutator, error) {
	c, ok := s.Mutators[name]
	if !ok {
		return nil, nil
	}
	return c, nil
}

// DeleteMutatorByName ...
func (s *FixtureStore) DeleteMutatorByName(name string) error {
	delete(s.Mutators, name)
	return nil
}

// UpdateMutator ...
func (s *FixtureStore) UpdateMutator(mutator *types.Mutator) error {
	s.Mutators[mutator.Name] = mutator
	return nil
}

// Checks

// GetChecks ...
func (s *FixtureStore) GetChecks() ([]*types.Check, error) {
	var checks []*types.Check
	for _, v := range s.Checks {
		checks = append(checks, v)
	}
	return checks, nil
}

// GetCheckByName ...
func (s *FixtureStore) GetCheckByName(name string) (*types.Check, error) {
	c, ok := s.Checks[name]
	if !ok {
		return nil, nil
	}
	return c, nil
}

// DeleteCheckByName ...
func (s *FixtureStore) DeleteCheckByName(name string) error {
	delete(s.Checks, name)
	return nil
}

// UpdateCheck ...
func (s *FixtureStore) UpdateCheck(check *types.Check) error {
	s.Checks[check.Name] = check
	return nil
}

// Events

// GetEvents ...
func (s *FixtureStore) GetEvents() ([]*types.Event, error) {
	return nil, nil
}

// GetEventsByEntity ...
func (s *FixtureStore) GetEventsByEntity(entityID string) ([]*types.Event, error) {
	return nil, nil
}

// GetEventByEntityCheck ...
func (s *FixtureStore) GetEventByEntityCheck(entityID, checkID string) (*types.Event, error) {
	return nil, nil
}

// UpdateEvent ...
func (s *FixtureStore) UpdateEvent(event *types.Event) error {
	return nil
}

// DeleteEventByEntityCheck ...
func (s *FixtureStore) DeleteEventByEntityCheck(entityID, checkID string) error {
	return nil
}

// NewFixtureStore returns a pointer to a new, initialized store. Each test
// requiring a store, should initialize its own so that tests can't
// pollute state.
func NewFixtureStore() *FixtureStore {
	s := &FixtureStore{
		Entities: map[string]*types.Entity{},
		Handlers: map[string]*types.Handler{},
		Mutators: map[string]*types.Mutator{},
		Checks:   map[string]*types.Check{},
	}

	for _, e := range entityFixtures {
		s.Entities[e.ID] = e
	}

	for _, h := range handlerFixtures {
		s.Handlers[h.Name] = h
	}

	for _, m := range mutatorFixtures {
		s.Mutators[m.Name] = m
	}

	for _, c := range checkFixtures {
		s.Checks[c.Name] = c
	}

	return s
}
