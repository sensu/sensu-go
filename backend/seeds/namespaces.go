package seeds

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func setupNamespaces(ctx context.Context, s storev2.NamespaceStore, config Config) error {
	namespaces := []*corev3.Namespace{
		defaultNamespace(),
	}

	for _, namespace := range namespaces {
		name := namespace.Metadata.Name

		if err := s.CreateIfNotExists(ctx, namespace); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := fmt.Sprintf("could not initialize the %s namespace", name)
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warnf("%s namespace already exists", name)
		}
	}

	return nil
}

func defaultNamespace() *corev3.Namespace {
	return &corev3.Namespace{
		Metadata: &corev2.ObjectMeta{
			Name:        "default",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
}
