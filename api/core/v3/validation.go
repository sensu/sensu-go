package v3

import (
	"errors"
	fmt "fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

func ValidateMetadata(meta *corev2.ObjectMeta) error {
	if meta == nil {
		return errors.New("nil metadata")
	}
	if err := corev2.ValidateName(meta.Name); err != nil {
		return fmt.Errorf("name %s", err)
	}
	if meta.Labels == nil {
		return errors.New("nil labels")
	}
	if meta.Annotations == nil {
		return errors.New("nil annotations")
	}
	return nil
}
