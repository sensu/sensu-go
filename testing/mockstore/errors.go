package mockstore

// DeleteError ...
func (s *MockStore) DeleteError(ctx context.Context, e, c, t string) error {
	args := s.Called(ctx, e, c, t)
	return args.Error(0)
}

// DeleteErrorsByEntityCheck ...
func (s *MockStore) DeleteErrorsByEntityCheck(ctx context.Context, e, c string) error {
	args := s.Called(ctx, e, c)
	return args.Error(0)
}

// DeleteErrorsByEntity ...
func (s *MockStore) DeleteErrorsByEntity(ctx context.Context, e string) error {
	s.DeleteErrorsByEntityCheck(ctx, e, "")
}

// GetErrors ...
func (s *MockStore) GetErrors(ctx context.Context) ([]*types.Error, error) {
	args := s.Called(ctx)
	return args.Get(0).([]*types.Error), args.Error(1)
}

// GetError ...
func (s *MockStore) GetError(ctx context.Context, e, c, t string) ([]*types.Error, error) {
	args := s.Called(ctx, e, c, t)
	return args.Get(0).([]*types.Error), args.Error(1)
}

// GetErrorsByEntity ...
func (s *MockStore) GetErrorsByEntity(ctx context.Context, entityID string) ([]*types.Error, error) {
	args := s.Called(ctx, entityID)
	return args.Get(0).([]*types.Error), args.Error(1)
}

// GetErrorByEntityCheck ...
func (s *MockStore) GetErrorByEntityCheck(ctx context.Context, entityID, checkID string) (*types.Error, error) {
	args := s.Called(ctx, entityID, checkID)
	return args.Get(0).(*types.Error), args.Error(1)
}

// CreateError ...
func (s *MockStore) CreateError(ctx context.Context, err *types.Error) error {
	args := s.Called(err)
	return args.Error(0)
}
