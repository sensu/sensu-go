package middlewares

import (
	"fmt"
	"math"
	"net/http"

	"github.com/sensu/sensu-go/backend/limiter"
)

// HeaderWarning is the header key for entity limit warnings
const HeaderWarning = "Warning"

// EntityLimiter is an HTTP middleware that surfaces warnings
// about entity limits.
type EntityLimiter struct {
	Limiter *limiter.EntityLimiter
}

// Then middleware
func (e EntityLimiter) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		e.LimitHeader(w)
		next.ServeHTTP(w, r)
	})
}

// LimitHeader calculates if the entity limit has been reached and appends a warning header.
func (e EntityLimiter) LimitHeader(w http.ResponseWriter) {
	entityLimit := e.Limiter.Limit()
	entityCount := math.Floor(average(e.Limiter.CountHistory()))
	if entityCount > float64(entityLimit) {
		w.Header().Set(HeaderWarning, fmt.Sprintf("You have exceeded the entity limit of %d for the free tier of Sensu Go: %d entities", entityLimit, int(entityCount)))
	}
}

func average(history []int) float64 {
	var total float64
	for _, value := range history {
		total += float64(value)
	}
	return total / float64(len(history))
}
