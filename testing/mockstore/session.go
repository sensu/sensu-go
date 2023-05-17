package mockstore

import "context"

func (s *MockStore) GetSession(ctx context.Context, username, sessionID string) (string, error) {
	args := s.Called(ctx, username, sessionID)
	return args.Get(0).(string), args.Error(1)
}

func (s *MockStore) UpdateSession(ctx context.Context, username, sessionID, state string) error {
	args := s.Called(ctx, username, sessionID, state)
	return args.Error(0)
}

func (s *MockStore) DeleteSession(ctx context.Context, username, sessionID string) error {
	args := s.Called(ctx, username, sessionID)
	return args.Error(0)
}
