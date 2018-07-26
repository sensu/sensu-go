package client

import (
	"encoding/json"

	"github.com/go-resty/resty"
)

type apiError struct {
	Message string `json:"error"`
	Code    uint32 `json:"code,omitempty"`
}

func (a apiError) Error() string {
	return a.Message
}

// UnmarshalError decode the API error
// TODO: Export err type from routers package.
func UnmarshalError(res *resty.Response) error {
	var apiErr apiError
	if err := json.Unmarshal(res.Body(), &apiErr); err != nil {
		apiErr.Message = string(res.Body())
	}
	return apiErr
}
