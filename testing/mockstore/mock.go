package mockstore

import (
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

// MockStore is a store used for testing. When using the MockStore in unit
// tests, stub out the behavior you wish to test against by assigning the
// appropriate function to the appropriate Func field. If you have forgotten
// to stub a particular function, the program will panic.
type MockStore struct {
	mock.Mock
}

//// Assets

// GetAssets ...
func (s *MockStore) GetAssets() ([]*types.Asset, error) {
	args := s.Called()
	return args.Get(0).([]*types.Asset), args.Error(1)
}

// GetAssetByName ...
func (s *MockStore) GetAssetByName(name string) (*types.Asset, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Asset), args.Error(1)
}

// DeleteAssetByName ...
func (s *MockStore) DeleteAssetByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateAsset ...
func (s *MockStore) UpdateAsset(asset *types.Asset) error {
	args := s.Called(asset)
	return args.Error(0)
}

//// Authentication

// CreateJWTSecret ...
func (s *MockStore) CreateJWTSecret(secret []byte) error {
	args := s.Called()
	return args.Error(0)
}

// GetJWTSecret ...
func (s *MockStore) GetJWTSecret() ([]byte, error) {
	args := s.Called()
	return []byte(args.String(0)), args.Error(1)
}

// UpdateJWTSecret ...
func (s *MockStore) UpdateJWTSecret(secret []byte) error {
	args := s.Called(secret)
	return args.Error(0)
}

//// CheckConfigs

// GetCheckConfigs ...
func (s *MockStore) GetCheckConfigs() ([]*types.CheckConfig, error) {
	args := s.Called()
	return args.Get(0).([]*types.CheckConfig), args.Error(1)
}

// GetCheckConfigByName ...
func (s *MockStore) GetCheckConfigByName(name string) (*types.CheckConfig, error) {
	args := s.Called(name)
	return args.Get(0).(*types.CheckConfig), args.Error(1)
}

// DeleteCheckConfigByName ...
func (s *MockStore) DeleteCheckConfigByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateCheckConfig ...
func (s *MockStore) UpdateCheckConfig(check *types.CheckConfig) error {
	args := s.Called(check)
	return args.Error(0)
}

//// Entities

// GetEntityByID ...
func (s *MockStore) GetEntityByID(id string) (*types.Entity, error) {
	args := s.Called(id)
	return args.Get(0).(*types.Entity), args.Error(1)
}

// UpdateEntity ...
func (s *MockStore) UpdateEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}

// DeleteEntity ...
func (s *MockStore) DeleteEntity(e *types.Entity) error {
	args := s.Called(e)
	return args.Error(0)
}

// DeleteEntityByID ...
func (s *MockStore) DeleteEntityByID(id string) error {
	args := s.Called(id)
	return args.Error(0)
}

// GetEntities ...
func (s *MockStore) GetEntities() ([]*types.Entity, error) {
	args := s.Called()
	return args.Get(0).([]*types.Entity), args.Error(1)
}

//// Events

// GetEvents ...
func (s *MockStore) GetEvents() ([]*types.Event, error) {
	args := s.Called()
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventsByEntity ...
func (s *MockStore) GetEventsByEntity(entityID string) ([]*types.Event, error) {
	args := s.Called(entityID)
	return args.Get(0).([]*types.Event), args.Error(1)
}

// GetEventByEntityCheck ...
func (s *MockStore) GetEventByEntityCheck(entityID, checkID string) (*types.Event, error) {
	args := s.Called(entityID, checkID)
	return args.Get(0).(*types.Event), args.Error(1)
}

// UpdateEvent ...
func (s *MockStore) UpdateEvent(event *types.Event) error {
	args := s.Called(event)
	return args.Error(0)
}

// DeleteEventByEntityCheck ...
func (s *MockStore) DeleteEventByEntityCheck(entityID, checkID string) error {
	args := s.Called(entityID, checkID)
	return args.Error(0)
}

//// Handlers

// GetHandlers ...
func (s *MockStore) GetHandlers() ([]*types.Handler, error) {
	args := s.Called()
	return args.Get(0).([]*types.Handler), args.Error(1)
}

// GetHandlerByName ...
func (s *MockStore) GetHandlerByName(name string) (*types.Handler, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Handler), args.Error(1)
}

// DeleteHandlerByName ...
func (s *MockStore) DeleteHandlerByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateHandler ...
func (s *MockStore) UpdateHandler(handler *types.Handler) error {
	args := s.Called(handler)
	return args.Error(0)
}

//// Keepalives

// UpdateKeepalive ...
func (s *MockStore) UpdateKeepalive(entityID string, expiration int64) error {
	args := s.Called(entityID, expiration)
	return args.Error(0)
}

// GetKeepalive ...
func (s *MockStore) GetKeepalive(entityID string) (int64, error) {
	args := s.Called(entityID)
	return args.Get(0).(int64), args.Error(1)
}

//// Mutators

// GetMutators ...
func (s *MockStore) GetMutators() ([]*types.Mutator, error) {
	args := s.Called()
	return args.Get(0).([]*types.Mutator), args.Error(1)
}

// GetMutatorByName ...
func (s *MockStore) GetMutatorByName(name string) (*types.Mutator, error) {
	args := s.Called(name)
	return args.Get(0).(*types.Mutator), args.Error(1)
}

// DeleteMutatorByName ...
func (s *MockStore) DeleteMutatorByName(name string) error {
	args := s.Called(name)
	return args.Error(0)
}

// UpdateMutator ...
func (s *MockStore) UpdateMutator(mutator *types.Mutator) error {
	args := s.Called(mutator)
	return args.Error(0)
}

//// Users

// CreateUser ...
func (s *MockStore) CreateUser(user *types.User) error {
	args := s.Called(user)
	return args.Error(0)
}

// DeleteUserByName ...
func (s *MockStore) DeleteUserByName(username string) error {
	args := s.Called(username)
	return args.Error(0)
}

// GetUser ...
func (s *MockStore) GetUser(username string) (*types.User, error) {
	args := s.Called(username)
	return args.Get(0).(*types.User), args.Error(1)
}

// GetUsers ...
func (s *MockStore) GetUsers() ([]*types.User, error) {
	args := s.Called()
	return args.Get(0).([]*types.User), args.Error(1)
}

// UpdateUser ...
func (s *MockStore) UpdateUser(user *types.User) error {
	args := s.Called(user)
	return args.Error(0)
}
