package seeds

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

type userFn func() (*corev2.User, error)

func setupUsers(ctx context.Context, s storev2.Interface, config Config) error {
	userFns := []userFn{
		adminUser(config.AdminUsername, config.AdminPassword),
		agentUser(),
	}

	for _, userFn := range userFns {
		user, err := userFn()
		fmt.Println(user, err)
		if err != nil {
			msg := "could not build user: %w"
			logger.WithError(err).Error(msg)
			return fmt.Errorf("%s: %w", msg, err)
		}

		name := user.Username

		if err := createResource(ctx, s, user); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := fmt.Sprintf("could not initialize the %s user", name)
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warnf("%s user already exists", name)
		}
	}

	return nil
}

func buildUser(username, password string, groups []string) (*corev2.User, error) {
	hash, err := bcrypt.HashPassword(password)
	if err != nil {
		return nil, err
	}

	return &corev2.User{
		Username:     username,
		Password:     hash,
		PasswordHash: hash,
		Groups:       groups,
	}, nil
}

func adminUser(username, password string) userFn {
	return func() (*corev2.User, error) {
		return buildUser(username, password, []string{"cluster-admins"})
	}
}

func agentUser() userFn {
	username := "agent"
	password := "P@ssw0rd!"
	groups := []string{"system:agents"}

	return func() (*corev2.User, error) {
		return buildUser(username, password, groups)
	}
}
