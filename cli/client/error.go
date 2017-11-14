package client

import (
	"encoding/json"
	"fmt"

	resty "gopkg.in/resty.v0"
)

type apiError struct {
	Message string `json:"error"`
	Code    uint32 `json:"code"`
}

func (a apiError) Error() string {
	return a.Message
}

// TODO: Export err type from routers package.
func unmarshalError(res *resty.Response) error {
	var apiErr apiError
	if err := json.Unmarshal(res.Body(), &apiErr); err != nil {
		return fmt.Errorf("unable to read error response")
	}
	return apiErr
}
