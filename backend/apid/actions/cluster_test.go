package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.etcd.io/etcd/client/v3"
	"golang.org/x/net/context"
)

type mockCluster struct {
}

func (mockCluster) MemberList(context.Context) (*clientv3.MemberListResponse, error) {
	return new(clientv3.MemberListResponse), nil
}

func (mockCluster) MemberAdd(context.Context, []string) (*clientv3.MemberAddResponse, error) {
	return new(clientv3.MemberAddResponse), nil
}

func (mockCluster) MemberRemove(context.Context, uint64) (*clientv3.MemberRemoveResponse, error) {
	return new(clientv3.MemberRemoveResponse), nil
}

func (mockCluster) MemberUpdate(context.Context, uint64, []string) (*clientv3.MemberUpdateResponse, error) {
	return new(clientv3.MemberUpdateResponse), nil
}

func (mockCluster) MemberPromote(context.Context, uint64) (*clientv3.MemberPromoteResponse, error) {
	return new(clientv3.MemberPromoteResponse), nil
}

func (mockCluster) MemberAddAsLearner(context.Context, []string) (*clientv3.MemberAddResponse, error) {
	return new(clientv3.MemberAddResponse), nil
}

var _ clientv3.Cluster = mockCluster{}

func TestMemberList(t *testing.T) {
	ctrl := NewClusterController(mockCluster{}, &mockstore.MockStore{})

	_, err := ctrl.MemberList(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberAdd(t *testing.T) {
	ctrl := NewClusterController(mockCluster{}, &mockstore.MockStore{})

	_, err := ctrl.MemberAdd(context.Background(), []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberUpdate(t *testing.T) {
	ctrl := NewClusterController(mockCluster{}, &mockstore.MockStore{})

	_, err := ctrl.MemberUpdate(context.Background(), 1234, []string{"foo"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestMemberRemove(t *testing.T) {
	ctrl := NewClusterController(mockCluster{}, &mockstore.MockStore{})

	_, err := ctrl.MemberRemove(context.Background(), 1234)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewClusterIDController(t *testing.T) {
	assert := assert.New(t)

	store := &mockstore.MockStore{}
	actions := NewClusterController(mockCluster{}, store)

	assert.NotNil(actions)
	assert.Equal(store, actions.store)
}

func TestGetClusterID(t *testing.T) {
	testCases := []struct {
		name            string
		ctx             context.Context
		storeErr        error
		expectedResult  string
		expectedErr     bool
		expectedErrCode ErrCode
	}{
		{
			name:           "Get",
			ctx:            context.Background(),
			expectedResult: uuid.New().String(),
		},
		{
			name:            "Not found",
			ctx:             context.Background(),
			storeErr:        &store.ErrNotFound{},
			expectedErr:     true,
			expectedErrCode: NotFound,
		},
		{
			name:            "Store error",
			ctx:             context.Background(),
			storeErr:        errors.New("some error"),
			expectedErr:     true,
			expectedErrCode: InternalErr,
		},
	}

	for _, tc := range testCases {
		store := &mockstore.MockStore{}
		actions := NewClusterController(mockCluster{}, store)

		t.Run(tc.name, func(t *testing.T) {
			assert := assert.New(t)

			store.
				On("GetClusterID", mock.Anything, mock.Anything).
				Return(tc.expectedResult, tc.storeErr)

			result, err := actions.ClusterID(tc.ctx)

			if tc.expectedErr {
				inferErr, ok := err.(Error)
				if ok {
					assert.Equal(tc.expectedErrCode, inferErr.Code)
				} else {
					assert.Error(err)
					assert.FailNow("Return value was not of type 'Error'")
				}
			} else {
				assert.NoError(err)
				assert.Equal(tc.expectedResult, result)
			}
		})
	}
}
