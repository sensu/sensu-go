package routers

import (
	"net/http"

	"github.com/sensu/sensu-go/backend/apid/actions"
)

// HandleGone returns a http.HandlerFunc that will respond with a HTTP 410 Gone
// status and will include the provided error message in the error body.
func HandleGone(message string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, _ *http.Request) {
		WriteError(w, actions.NewErrorf(actions.Gone, message))
	}
}
