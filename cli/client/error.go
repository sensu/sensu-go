package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-resty/resty"
	"github.com/sensu/sensu-go/backend/apid/actions"
)

// APIError describes an error message returned by the REST API
type APIError struct {
	Message string `json:"message"`
	Code    uint32 `json:"code,omitempty"`
}

func (a APIError) Error() string {
	return a.Message
}

// UnmarshalError decode the API error
// TODO: Export err type from routers package.
func UnmarshalError(res *resty.Response) error {
	var apiErr APIError

	switch res.StatusCode() {
	case http.StatusPaymentRequired:
		apiErr.Code = uint32(actions.PaymentRequired)
		apiErr.Message = "This functionality requires a valid Sensu Go license. Please install a valid license file and restart or contact Sales for a trial."

	default:
		if err := json.Unmarshal(res.Body(), &apiErr); err != nil {
			if len(res.Body()) > 0 {
				apiErr.Message = string(res.Body())
			} else {
				apiErr.Message = fmt.Sprintf("the API returned: %s", res.Status())
			}
		}
	}

	return apiErr
}
