package seeds

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func setupAPIKeys(ctx context.Context, s storev2.Interface, config Config) error {
	apiKeys := []*corev2.APIKey{}

	if config.AdminAPIKey != "" {
		apiKey := adminAPIKey(config.AdminUsername, config.AdminAPIKey)
		apiKeys = append(apiKeys, apiKey)
	}

	for _, apiKey := range apiKeys {
		name := apiKey.ObjectMeta.Name

		if err := createResource(ctx, s, apiKey); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := fmt.Sprintf("could not initialize the %s api key", name)
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warnf("%s api key already exists", name)
		}
	}

	return nil
}

func adminAPIKey(username, apiKey string) *corev2.APIKey {
	return &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name:      apiKey,
			CreatedBy: username,
		},
		Username:  username,
		CreatedAt: time.Now().Unix(),
	}
}
