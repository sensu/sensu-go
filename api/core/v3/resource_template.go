package v3

import corev2 "github.com/sensu/sensu-go/api/core/v2"

type ResourceTemplate struct {
	Metadata   *corev2.ObjectMeta `json:"metadata"`
	APIVersion string             `json:"api_version"`
	Type       string             `json:"type"`
	Template   string             `json:"template"`
	Path       string             `json:"path"`
}

func (r *ResourceTemplate) GetMetadata() *corev2.ObjectMeta {
	return r.Metadata
}
