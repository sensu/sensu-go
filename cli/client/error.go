package client

import (
	"encoding/json"
	"errors"

	resty "gopkg.in/resty.v0"
)

// TODO: Export err type from routers package.
func unmarshalError(res *resty.Response) error {
	err := map[string]interface{}{}
	json.Unmarshal(res.Body(), &err)

	if msg, ok := err["error"].(string); ok {
		return errors.New(msg)
	}
	return errors.New("unable to read error response")
}
