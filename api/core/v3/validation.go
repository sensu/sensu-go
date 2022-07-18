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

// ValidateGlobalMetadata validates ObjectMeta for global (unnamespaced)
// resources. To be used on Resources implementing GlobalResource
func ValidateGlobalMetadata(meta *corev2.ObjectMeta) error {
	if meta == nil {
		return errors.New("nil metadata")
	}
	if meta.Namespace != "" {
		return fmt.Errorf(
			"global resources must have empty namesapce: got %s",
			meta.Namespace,
		)
	}
	return nil
}
