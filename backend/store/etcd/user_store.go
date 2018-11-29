package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/types"
)

func getUserPath(id string) string {
	return fmt.Sprintf("%s/users/%s", EtcdRoot, id)
}

// AuthenticateUser authenticates a User by username and password.
func (s *Store) AuthenticateUser(ctx context.Context, username, password string) (*types.User, error) {
	user, err := s.GetUser(ctx, username)
	if user == nil {
		return nil, fmt.Errorf("User %s does not exist", username)
	} else if err != nil {
		return nil, err
	}

	if user.Disabled {
		return nil, fmt.Errorf("User %s is disabled", username)
	}

	ok := bcrypt.CheckPassword(user.Password, password)
	if !ok {
		return nil, fmt.Errorf("Wrong password for user %s", username)
	}

	return user, nil
}

// CreateUser creates a new user
func (s *Store) CreateUser(u *types.User) error {
	userBytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	// We need to prepare a transaction to verify the version of the key
	// corresponding to the user in etcd in order to ensure we only put the key
	// if it does not exist
	cmp := clientv3.Compare(clientv3.Version(getUserPath(u.Username)), "=", 0)
	req := clientv3.OpPut(getUserPath(u.Username), string(userBytes))
	res, err := s.client.Txn(context.TODO()).If(cmp).Then(req).Commit()
	if err != nil {
		return err
	}
	if !res.Succeeded {
		return fmt.Errorf("user %s already exists", u.Username)
	}

	return nil
}

// DeleteUser deletes a User.
// NOTE:
// Store probably shouldn't be responsible for deleting the token;
// business logic.
func (s *Store) DeleteUser(ctx context.Context, user *types.User) error {
	// Mark it as disabled
	user.Disabled = true

	// Marshal the user struct
	userBytes, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Construct the list of operations to make in the transaction
	userKey := getUserPath(user.Username)
	txn := s.client.Txn(ctx).
		// Ensure that the key exists
		If(clientv3.Compare(clientv3.CreateRevision(userKey), ">", 0)).
		// If key exists, delete user & any access token from allow list
		Then(
			clientv3.OpPut(userKey, string(userBytes)),
			clientv3.OpDelete(
				getTokenPath(user.Username, ""),
				clientv3.WithPrefix(),
			),
		)

	res, serr := txn.Commit()
	if serr != nil {
		return err
	}

	if !res.Succeeded {
		logger.
			WithField("username", user.Username).
			Info("given user was not already persisted")
	}

	return nil
}

// GetUser gets a User.
func (s *Store) GetUser(ctx context.Context, username string) (*types.User, error) {
	resp, err := s.client.Get(ctx, getUserPath(username), clientv3.WithLimit(1))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, nil
	}

	user := &types.User{}
	err = json.Unmarshal(resp.Kvs[0].Value, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUsers retrieves all enabled users
func (s *Store) GetUsers() ([]*types.User, error) {
	allUsers, err := s.GetAllUsers()
	if err != nil {
		return allUsers, err
	}

	var users []*types.User
	for _, user := range allUsers {
		// Verify that the user is not disabled
		if !user.Disabled {
			users = append(users, user)
		}
	}

	return users, nil
}

// GetAllUsers retrieves all users
func (s *Store) GetAllUsers() ([]*types.User, error) {
	resp, err := s.client.Get(context.TODO(), getUserPath(""), clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) == 0 {
		return []*types.User{}, nil
	}

	usersArray := []*types.User{}
	for _, kv := range resp.Kvs {
		user := &types.User{}
		err = json.Unmarshal(kv.Value, user)
		if err != nil {
			return nil, err
		}

		usersArray = append(usersArray, user)
	}

	return usersArray, nil
}

// UpdateUser updates a User.
func (s *Store) UpdateUser(u *types.User) error {
	bytes, err := json.Marshal(u)
	if err != nil {
		return err
	}

	_, err = s.client.Put(context.TODO(), getUserPath(u.Username), string(bytes))
	return err
}
