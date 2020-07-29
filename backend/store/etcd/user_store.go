package etcd

import (
	"context"
	"fmt"
	"path"

	"github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/store"
)

const usersPathPrefix = "users"

func getUserPath(id string) string {
	return path.Join(store.Root, usersPathPrefix, id)
}

// GetUsersPath gets the path of the user store.
func GetUsersPath(ctx context.Context, id string) string {
	return path.Join(store.Root, usersPathPrefix, id)
}

// AuthenticateUser authenticates a User by username and password.
func (s *Store) AuthenticateUser(ctx context.Context, username, password string) (*corev2.User, error) {
	user, err := s.GetUser(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, &store.ErrNotFound{Key: username}
	}

	if user.Disabled {
		return nil, &store.ErrNotValid{Err: fmt.Errorf("user %s is disabled", username)}
	}

	// Check if we have an explicitly hashed password, otherwise fallback to the
	// password field for backward compatiblility
	passwordHash := user.PasswordHash
	if passwordHash == "" {
		passwordHash = user.Password
	}
	ok := bcrypt.CheckPassword(passwordHash, password)
	if !ok {
		return nil, &store.ErrNotValid{Err: fmt.Errorf("wrong password for user %s", username)}
	}

	return user, nil
}

// CreateUser creates a new user
func (s *Store) CreateUser(u *corev2.User) error {
	userBytes, err := proto.Marshal(u)
	if err != nil {
		return &store.ErrNotValid{Err: err}
	}

	// We need to prepare a transaction to verify the version of the key
	// corresponding to the user in etcd in order to ensure we only put the key
	// if it does not exist
	cmp := clientv3.Compare(clientv3.Version(getUserPath(u.Username)), "=", 0)
	req := clientv3.OpPut(getUserPath(u.Username), string(userBytes))
	var res *clientv3.TxnResponse
	err = Backoff(context.TODO()).Retry(func(n int) (done bool, err error) {
		res, err = s.client.Txn(context.TODO()).If(cmp).Then(req).Commit()
		return RetryRequest(n, err)
	})
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return &store.ErrAlreadyExists{Key: u.Username}
	}

	return nil
}

// GetUser gets a User.
func (s *Store) GetUser(ctx context.Context, username string) (*corev2.User, error) {
	var user corev2.User
	err := Get(ctx, s.client, getUserPath(username), &user)
	if err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			err = nil
		}
		return nil, err
	}
	return &user, nil
}

// GetUsers retrieves all enabled users
func (s *Store) GetUsers() ([]*corev2.User, error) {
	allUsers, err := s.GetAllUsers(&store.SelectionPredicate{})
	if err != nil {
		return allUsers, err
	}

	var users []*corev2.User
	for _, user := range allUsers {
		// Verify that the user is not disabled
		if !user.Disabled {
			users = append(users, user)
		}
	}

	return users, nil
}

// GetAllUsers retrieves all users
func (s *Store) GetAllUsers(pred *store.SelectionPredicate) ([]*corev2.User, error) {
	users := []*corev2.User{}
	err := List(context.Background(), s.client, GetUsersPath, &users, pred)
	return users, err
}

// UpdateUser updates a User.
func (s *Store) UpdateUser(u *corev2.User) error {
	bytes, err := proto.Marshal(u)
	if err != nil {
		return &store.ErrEncode{Err: err}
	}

	return Backoff(context.TODO()).Retry(func(n int) (done bool, err error) {
		_, err = s.client.Put(context.TODO(), getUserPath(u.Username), string(bytes))
		return RetryRequest(n, err)
	})
}
