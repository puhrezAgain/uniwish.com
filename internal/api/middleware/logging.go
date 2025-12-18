/*
uniwish.com/interal/api/middleware/logging

contains our logging middleware, wrapping http.Response as StatusWriter to capture status code, along other info
*/
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type StatusWriter struct {
	http.ResponseWriter
	status int
}

func NewStatusWriter(w http.ResponseWriter) *StatusWriter {
	return &StatusWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}
func (sw *StatusWriter) WriteHeader(code int) {
	sw.status = code
	sw.ResponseWriter.WriteHeader(code)
}

func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sw := NewStatusWriter(w)
			start := time.Now()
			next.ServeHTTP(sw, r)

			logger.Info(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", sw.status,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
