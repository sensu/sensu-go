package client

import (
	"encoding/json"

	"github.com/go-resty/resty"
)

// APIError describes an error message returned by the REST API
type APIError struct {
	Message string `json:"error"`
	Code    uint32 `json:"code,omitempty"`
}

func (a APIError) Error() string {
	return a.Message
}

// UnmarshalError decode the API error
// TODO: Export err type from routers package.
func UnmarshalError(res *resty.Response) error {
	var apiErr APIError
	if err := json.Unmarshal(res.Body(), &apiErr); err != nil {
		apiErr.Message = string(res.Body())
	}
	return apiErr
}
