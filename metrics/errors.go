package metrics

import (
	"fmt"
)

func FormatRegistrationErr(metric string, err error) error {
	return fmt.Errorf("unable to register %s metric: %w", metric, err)
}
