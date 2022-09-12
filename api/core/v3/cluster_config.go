package v3

import corev2 "github.com/sensu/sensu-go/api/core/v2"

type ClusterConfig struct {
	Metadata *corev2.ObjectMeta
}

func (c *ClusterConfig) GetMetadata() *corev2.ObjectMeta {
	return c.Metadata
}
