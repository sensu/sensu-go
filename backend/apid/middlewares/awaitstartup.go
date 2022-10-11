package middlewares

import (
	"fmt"
	"net/http"
	"sync"
)

type AwaitStartupMiddleware struct {
	RetryAfterSeconds int
	ResponseText      string
	isReady           bool
	mu                sync.RWMutex
}

func (m *AwaitStartupMiddleware) Ready() {
	m.mu.Lock()
	m.isReady = true
	m.mu.Unlock()
}

func (m *AwaitStartupMiddleware) Then(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.mu.RLock()
		ready := m.isReady
		m.mu.RUnlock()
		if !ready {
			Logger.Info("apid temporarily unavailable during startup")
			if m.RetryAfterSeconds > 0 {
				w.Header().Set("Retry-After", fmt.Sprint(m.RetryAfterSeconds))
			}
			text := http.StatusText(http.StatusServiceUnavailable)
			if m.ResponseText != "" {
				text = m.ResponseText
			}
			http.Error(w, text, http.StatusServiceUnavailable)
			return
		}
		next.ServeHTTP(w, r)
	})
}
