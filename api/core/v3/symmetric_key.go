package v3

import corev2 "github.com/sensu/sensu-go/api/core/v2"

type SymmetricKey struct {
	Metadata *corev2.ObjectMeta `json:"metadata"`
	Value    []byte             `json:"value"`
}

func (s *SymmetricKey) GetMetadata() *corev2.ObjectMeta {
	return s.Metadata
}
