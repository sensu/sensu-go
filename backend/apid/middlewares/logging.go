package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

// SimpleLogger log request path and duration
type SimpleLogger struct{}

// Then middleware
func (m SimpleLogger) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := float64(time.Now().Sub(start)) / float64(time.Millisecond)

		logger.WithFields(logrus.Fields{
			"duration": fmt.Sprintf("%.3fms", duration),
			"path":     r.URL.Path,
		}).Info()
	})
}
