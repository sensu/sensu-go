package middlewares

import (
	"net/http"
	"strings"

	"github.com/sensu/sensu-go/cli/commands/helpers"
)

// EntityLimiter is an HTTP middleware that errors if the entity
// limit has been reached.
type EntityLimiter struct{}

// Then middleware
func (e EntityLimiter) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get(helpers.HeaderWarning)
		if strings.Contains(header, "We noticed you've reached") {
			logger.Error("entity limit has been reached")
			http.Error(w, "Payment required", http.StatusPaymentRequired)
			return
		}
		next.ServeHTTP(w, r)
	})
}
