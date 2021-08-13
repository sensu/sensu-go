package pipeline

import (
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
)

// Config holds the configuration for a Pipeline.
type Config struct {
	Store                  store.Store
	AssetGetter            asset.Getter
	BackendEntity          *corev2.Entity
	StoreTimeout           time.Duration
	SecretsProviderManager *secrets.ProviderManager
	LicenseGetter          licensing.Getter
}

// Option is a functional option used to configure Pipelines.
type Option func(*Pipeline)
