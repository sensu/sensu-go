package schedulerd

import (
	corev3 "github.com/sensu/core/v3"
	cachev2 "github.com/sensu/sensu-go/backend/store/cache/v2"
)

type EntityCache = *cachev2.Resource[*corev3.EntityConfig, corev3.EntityConfig]
type EntityCacheValue = cachev2.Value[*corev3.EntityConfig, corev3.EntityConfig]
