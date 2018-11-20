package graphql

import "github.com/sensu/sensu-go/backend/apid/actions"

// TODO Remove me; unused
func handleControllerResults(record interface{}, err error) (interface{}, error) {
	// If no error occurred simply return the record
	if err == nil {
		return record, nil
	}

	// If the controller returned a not found error, swallow it and return nil
	s, ok := actions.StatusFromError(err)
	if ok && s == actions.NotFound {
		return nil, nil
	}

	// Avoid inadvertantly leaking record when an error occurs
	return nil, err
}
