package middlewares

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

// SimpleLogger log request path and duration
type SimpleLogger struct{}

// Then middleware
func (m SimpleLogger) Then(next http.Handler) http.Handler {
	if logger.Logger.Level < logrus.InfoLevel {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writerWithCapture := makeResponseWriterWithCapture(w)
		next.ServeHTTP(writerWithCapture, r)

		duration := float64(time.Since(start)) / float64(time.Millisecond)
		logEntry := logger.WithFields(logrus.Fields{
			"duration": fmt.Sprintf("%.3fms", duration),
			"status":   writerWithCapture.Status(),
			"size":     writerWithCapture.Size(),
			"path":     r.URL.Path,
			"method":   r.Method,
		})
		logEntry.Info("request completed")
	})
}

//
// responseLogger is borrowed verbatim from github.com/gorilla/handlers.

func makeResponseWriterWithCapture(w http.ResponseWriter) loggingResponseWriter {
	var logger loggingResponseWriter = &responseLogger{w: w, status: http.StatusOK}
	if _, ok := w.(http.Hijacker); ok {
		logger = &hijackLogger{responseLogger{w: w, status: http.StatusOK}}
	}
	h, ok1 := logger.(http.Hijacker)
	c, ok2 := w.(http.CloseNotifier)
	if ok1 && ok2 {
		return hijackCloseNotifier{logger, h, c}
	}
	if ok2 {
		return &closeNotifyWriter{logger, c}
	}
	return logger
}

type loggingResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	Status() int
	Size() int
}

// responseLogger is wrapper of http.ResponseWriter that keeps track of its HTTP
// status code and body size
type responseLogger struct {
	w      http.ResponseWriter
	status int
	size   int
}

func (l *responseLogger) Header() http.Header {
	return l.w.Header()
}

func (l *responseLogger) Write(b []byte) (int, error) {
	size, err := l.w.Write(b)
	l.size += size
	return size, err
}

func (l *responseLogger) WriteHeader(s int) {
	l.w.WriteHeader(s)
	l.status = s
}

func (l *responseLogger) Status() int {
	return l.status
}

func (l *responseLogger) Size() int {
	return l.size
}

func (l *responseLogger) Flush() {
	f, ok := l.w.(http.Flusher)
	if ok {
		f.Flush()
	}
}

type hijackLogger struct {
	responseLogger
}

func (l *hijackLogger) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h := l.responseLogger.w.(http.Hijacker)
	conn, rw, err := h.Hijack()
	if err == nil && l.responseLogger.status == 0 {
		// The status will be StatusSwitchingProtocols if there was no error and
		// WriteHeader has not been called yet
		l.responseLogger.status = http.StatusSwitchingProtocols
	}
	return conn, rw, err
}

type closeNotifyWriter struct {
	loggingResponseWriter
	http.CloseNotifier
}

type hijackCloseNotifier struct {
	loggingResponseWriter
	http.Hijacker
	http.CloseNotifier
}
