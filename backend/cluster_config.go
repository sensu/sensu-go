package backend

import (
	"context"

	"github.com/google/uuid"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func GetClusterID(ctx context.Context, s storev2.Interface) (string, error) {
	// first try to create a new cluster ID
	clusterID := uuid.New().String()
	clusterConfig := &corev3.ClusterConfig{
		Metadata: &corev2.ObjectMeta{
			Name:        clusterID,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
	}
	ccStore := storev2.Of[*corev3.ClusterConfig](s)
	err := ccStore.CreateIfNotExists(ctx, clusterConfig)
	if err == nil {
		return clusterID, nil
	}
	if _, ok := err.(*store.ErrAlreadyExists); ok {
		clusterConfigs, err := ccStore.List(ctx, storev2.ID{}, nil)
		if err != nil {
			return "", err
		}
		if len(clusterConfigs) == 0 {
			panic("fatal error")
		}
		return clusterConfigs[0].Metadata.Name, nil
	} else {
		return "", err
	}
}
