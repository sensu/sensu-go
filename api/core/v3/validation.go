package v3

import (
	"errors"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func ValidateMetadata(meta *corev2.ObjectMeta) error {
	if meta == nil {
		return errors.New("nil metadata")
	}
	if err := corev2.ValidateName(meta.Name); err != nil {
		return err
	}
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	return nil
}
